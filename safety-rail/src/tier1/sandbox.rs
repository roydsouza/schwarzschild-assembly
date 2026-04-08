use crate::{ExecutionErrorKind, ExecutionResult, VerifiedArtifact, ViolationReport};
use std::collections::HashMap;
use std::time::{Duration, Instant};
use wasmtime::{Config, Engine, Linker, Module, Store, ResourceLimiter};
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiView, Table};

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
        let artifact_id = artifact.proposal.id;

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
             if let Ok(dir) = wasmtime_wasi::Dir::open_ambient_dir(&root_path, wasmtime_wasi::CapFlags::READ) {
                 // WASI p2 preopen logic (simplified for Tier 1)
                 // wasi_builder.preopened_dir(dir, ".").expect("Preopen root");
             }
        }

        // Target path is writable if provided
        if let Some(target) = &artifact.proposal.target_path {
            // wasi_builder.preopened_dir(..., target).expect("Preopen target");
        }

        let wasi = wasi_builder.build();
        let table = Table::new();

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
        store.out_of_fuel_trap();
        
        // Start timeout monitor (epoch interruption)
        let engine_clone = self.engine.clone();
        std::thread::spawn(move || {
            std::thread::sleep(timeout);
            engine_clone.increment_epoch();
        });

        // 3. Compile and Link
        // For Tier 1, we assume the proposal payload contains raw WASM bytes if operation is ExecuteCode
        let wasm_bytes = &artifact.proposal.payload;
        let module = match Module::new(&self.engine, wasm_bytes) {
            Ok(m) => m,
            Err(e) => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::CompilationError,
                message: format!("WASM compilation failed: {}", e),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        let mut linker = Linker::new(&self.engine);
        wasmtime_wasi::add_to_linker_sync(&mut linker, |s| s).expect("Add WASI to linker");

        // 4. Instantiate and Run
        let instance = match linker.instantiate(&mut store, &module) {
            Ok(inst) => inst,
            Err(e) => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::InstantiationError,
                message: format!("WASM instantiation failed: {}", e),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        let func = match instance.get_typed_func::<(), ()>(&mut store, "main") {
            Ok(f) => f,
            Err(e) => return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: ExecutionErrorKind::InstantiationError, 
                message: format!("WASM main function missing: {}", e),
                exit_code: None,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
            },
        };

        match func.call(&mut store, ()) {
            Ok(_) => ExecutionResult::Success {
                artifact_id: artifact.proposal.id,
                output: Vec::new(), // In Tier 1, we don't capture stdout bytes yet
                exit_code: 0,
                elapsed_ms: start_time.elapsed().as_millis() as u64,
                peak_memory_bytes: 0, // Placeholder
            },
            Err(e) => {
                let error_kind = if e.to_string().contains("epoch") {
                    ExecutionErrorKind::Timeout
                } else if e.to_string().contains("memory") {
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
    table: Table,
    limits: SandboxLimits,
}

impl WasiView for SandboxState {
    fn table(&self) -> &Table {
        &self.table
    }
    fn table_mut(&mut self) -> &mut Table {
        &mut self.table
    }
    fn ctx(&self) -> &WasiCtx {
        &self.wasi
    }
    fn ctx_mut(&mut self) -> &mut WasiCtx {
        &mut self.wasi
    }
}

impl ResourceLimiter for SandboxState {
    fn memory_growing(&mut self, current: usize, desired: usize, _maximum: Option<usize>) -> Result<bool, wasmtime::Error> {
        if desired > self.limits.memory {
            return Ok(false);
        }
        Ok(true)
    }

    fn table_growing(&mut self, _current: u32, _desired: u32, _maximum: Option<u32>) -> Result<bool, wasmtime::Error> {
        Ok(true)
    }
}
