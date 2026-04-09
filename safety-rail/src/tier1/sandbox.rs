use crate::{ExecutionErrorKind, ExecutionResult, VerifiedArtifact};
use std::time::{Duration, Instant};
use wasmtime::{Config, Engine, Store, ResourceLimiter};
use wasmtime::component::{Component, Linker};
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiView, ResourceTable};

/// Implementation of the WASM sandbox using Wasmtime.
pub(crate) struct WasmSandbox {
    engine: Engine,
}

impl WasmSandbox {
    pub fn new() -> Result<Self, String> {
        let mut config = Config::new();
        config.wasm_component_model(true);
        config.epoch_interruption(true);
        config.consume_fuel(true);

        let engine = Engine::new(&config).map_err(|e| format!("Failed to create Wasmtime engine: {}", e))?;
        Ok(Self { engine })
    }

    pub fn execute(&self, artifact: &VerifiedArtifact) -> ExecutionResult {
        let start_time = Instant::now();

        // Configuration
        let memory_limit = 256 * 1024 * 1024; // 256 MiB
        let timeout = Duration::from_secs(5);
        let fuel_limit = 100_000_000; // Arbitrary fuel limit for smoke test

        // 1. Setup WASI context
        let mut wasi_builder = WasiCtxBuilder::new();
        wasi_builder.inherit_stdout().inherit_stderr();

        // Dual-Path Preopen Strategy
        // Root is read-only
        if let Ok(root_path) = std::env::current_dir() {
            let path_str = root_path.to_string_lossy().to_string();
            let _ = wasi_builder.preopened_dir(&path_str, ".", wasmtime_wasi::DirPerms::READ, wasmtime_wasi::FilePerms::READ);
        }

        // Target path is writable if provided
        if let Some(target) = &artifact.proposal.target_path {
             let _ = wasi_builder.preopened_dir(target, target, wasmtime_wasi::DirPerms::all(), wasmtime_wasi::FilePerms::all());
        }

        let wasi = wasi_builder.build();
        let table = ResourceTable::new();

        let mut store = Store::new(
            &self.engine,
            SandboxState {
                wasi,
                table,
                limits: SandboxLimits {
                    memory: memory_limit,
                },
            },
        );

        // 2. Set limits
        store.set_fuel(fuel_limit).unwrap();
        store.set_epoch_deadline(1); // Trap on the first epoch increment
        
        // Start timeout monitor (epoch interruption)
        let engine_clone = self.engine.clone();
        std::thread::spawn(move || {
            std::thread::sleep(timeout);
            engine_clone.increment_epoch();
        });

        // 3. Compile and Link
        // For Tier 1, we assume the proposal payload contains raw component bytes if operation is ExecuteCode
        let wasm_bytes = &artifact.proposal.payload;
        let component = match Component::new(&self.engine, wasm_bytes) {
            Ok(c) => c,
            Err(e) => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::CompilationError,
                message: format!("WASM component compilation failed: {}", e),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        let mut linker = Linker::new(&self.engine);
        wasmtime_wasi::add_to_linker_sync(&mut linker).expect("Add WASI to linker");

        // 4. Instantiate and Run
        let instance = match linker.instantiate(&mut store, &component) {
            Ok(inst) => inst,
            Err(e) => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::InstantiationError,
                message: format!("WASM instantiation failed: {}", e),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        // For components, we often use exports from the root
        // Here we assume a simple 'run' function or similar convention for Tier 1
        let func = match instance.get_func(&mut store, "main") {
            Some(f) => f,
            None => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::InstantiationError, 
                message: "WASM component 'main' function missing".to_string(),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        match func.call(&mut store, &[], &mut []) {
            Ok(_) => ExecutionResult::Success {
                artifact_id: artifact.proposal.id,
                output: Vec::new(), 
                exit_code: 0,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
                peak_memory_bytes: 0, 
            },
            Err(e) => {
                let err_msg = format!("{:#}", e).to_lowercase();
                // Check for interruption/timeout/fuel
                let error_kind = if err_msg.contains("epoch") || err_msg.contains("interrupt") || err_msg.contains("timeout") || err_msg.contains("fuel") {
                    ExecutionErrorKind::Timeout
                } else if err_msg.contains("memory") || err_msg.contains("resource limit") || err_msg.contains("exhausted") {
                    ExecutionErrorKind::MemoryExhausted
                } else {
                    ExecutionErrorKind::RuntimeTrap
                };

                ExecutionResult::Failure {
                    artifact_id: artifact.proposal.id,
                    error_kind,
                    message: format!("WASM execution failed: {}", e),
                    exit_code: None,
                    elapsed_ms: start_time.elapsed().as_millis() as u64,
                }
            }
        }
    }
}

struct SandboxLimits {
    memory: usize,
}

struct SandboxState {
    wasi: WasiCtx,
    table: ResourceTable,
    limits: SandboxLimits,
}

impl WasiView for SandboxState {
    fn table(&mut self) -> &mut ResourceTable {
        &mut self.table
    }
    fn ctx(&mut self) -> &mut WasiCtx {
        &mut self.wasi
    }
}

impl ResourceLimiter for SandboxState {
    fn memory_growing(&mut self, _current: usize, desired: usize, _maximum: Option<usize>) -> Result<bool, wasmtime::Error> {
        if desired > self.limits.memory {
            return Ok(false);
        }
        Ok(true)
    }

    fn table_growing(&mut self, _current: u32, _desired: u32, _maximum: Option<u32>) -> Result<bool, wasmtime::Error> {
        Ok(true)
    }
}
