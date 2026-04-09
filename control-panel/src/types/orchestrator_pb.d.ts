import * as jspb from 'google-protobuf'



export class Empty extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Empty.AsObject;
  static toObject(includeInstance: boolean, msg: Empty): Empty.AsObject;
  static serializeBinaryToWriter(message: Empty, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Empty;
  static deserializeBinaryFromReader(message: Empty, reader: jspb.BinaryReader): Empty;
}

export namespace Empty {
  export type AsObject = {
  }
}

export class OperationStatus extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): OperationStatus;

  getErrorCode(): string;
  setErrorCode(value: string): OperationStatus;

  getMessage(): string;
  setMessage(value: string): OperationStatus;

  getRequestId(): string;
  setRequestId(value: string): OperationStatus;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OperationStatus.AsObject;
  static toObject(includeInstance: boolean, msg: OperationStatus): OperationStatus.AsObject;
  static serializeBinaryToWriter(message: OperationStatus, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OperationStatus;
  static deserializeBinaryFromReader(message: OperationStatus, reader: jspb.BinaryReader): OperationStatus;
}

export namespace OperationStatus {
  export type AsObject = {
    success: boolean,
    errorCode: string,
    message: string,
    requestId: string,
  }
}

export class FactoryRequest extends jspb.Message {
  getFactoryType(): string;
  setFactoryType(value: string): FactoryRequest;

  getFactoryName(): string;
  setFactoryName(value: string): FactoryRequest;

  getConfigJson(): Uint8Array | string;
  getConfigJson_asU8(): Uint8Array;
  getConfigJson_asB64(): string;
  setConfigJson(value: Uint8Array | string): FactoryRequest;

  getDomainMetrics(): DomainFitnessExtension | undefined;
  setDomainMetrics(value?: DomainFitnessExtension): FactoryRequest;
  hasDomainMetrics(): boolean;
  clearDomainMetrics(): FactoryRequest;

  getRequestId(): string;
  setRequestId(value: string): FactoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryRequest): FactoryRequest.AsObject;
  static serializeBinaryToWriter(message: FactoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryRequest;
  static deserializeBinaryFromReader(message: FactoryRequest, reader: jspb.BinaryReader): FactoryRequest;
}

export namespace FactoryRequest {
  export type AsObject = {
    factoryType: string,
    factoryName: string,
    configJson: Uint8Array | string,
    domainMetrics?: DomainFitnessExtension.AsObject,
    requestId: string,
  }
}

export class FactoryResponse extends jspb.Message {
  getFactoryId(): FactoryID | undefined;
  setFactoryId(value?: FactoryID): FactoryResponse;
  hasFactoryId(): boolean;
  clearFactoryId(): FactoryResponse;

  getStatus(): FactoryStatus | undefined;
  setStatus(value?: FactoryStatus): FactoryResponse;
  hasStatus(): boolean;
  clearStatus(): FactoryResponse;

  getFitnessBaseline(): FitnessSnapshot | undefined;
  setFitnessBaseline(value?: FitnessSnapshot): FactoryResponse;
  hasFitnessBaseline(): boolean;
  clearFitnessBaseline(): FactoryResponse;

  getMerkleLeafHash(): string;
  setMerkleLeafHash(value: string): FactoryResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryResponse): FactoryResponse.AsObject;
  static serializeBinaryToWriter(message: FactoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryResponse;
  static deserializeBinaryFromReader(message: FactoryResponse, reader: jspb.BinaryReader): FactoryResponse;
}

export namespace FactoryResponse {
  export type AsObject = {
    factoryId?: FactoryID.AsObject,
    status?: FactoryStatus.AsObject,
    fitnessBaseline?: FitnessSnapshot.AsObject,
    merkleLeafHash: string,
  }
}

export class FactoryID extends jspb.Message {
  getId(): string;
  setId(value: string): FactoryID;

  getName(): string;
  setName(value: string): FactoryID;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryID.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryID): FactoryID.AsObject;
  static serializeBinaryToWriter(message: FactoryID, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryID;
  static deserializeBinaryFromReader(message: FactoryID, reader: jspb.BinaryReader): FactoryID;
}

export namespace FactoryID {
  export type AsObject = {
    id: string,
    name: string,
  }
}

export class FactoryStatus extends jspb.Message {
  getState(): FactoryState;
  setState(value: FactoryState): FactoryStatus;

  getLastHeartbeatMs(): number;
  setLastHeartbeatMs(value: number): FactoryStatus;

  getProposalsInFlight(): number;
  setProposalsInFlight(value: number): FactoryStatus;

  getErrorMessage(): string;
  setErrorMessage(value: string): FactoryStatus;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryStatus.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryStatus): FactoryStatus.AsObject;
  static serializeBinaryToWriter(message: FactoryStatus, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryStatus;
  static deserializeBinaryFromReader(message: FactoryStatus, reader: jspb.BinaryReader): FactoryStatus;
}

export namespace FactoryStatus {
  export type AsObject = {
    state: FactoryState,
    lastHeartbeatMs: number,
    proposalsInFlight: number,
    errorMessage: string,
  }
}

export class FactoryList extends jspb.Message {
  getFactoriesList(): Array<FactoryEntry>;
  setFactoriesList(value: Array<FactoryEntry>): FactoryList;
  clearFactoriesList(): FactoryList;
  addFactories(value?: FactoryEntry, index?: number): FactoryEntry;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryList.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryList): FactoryList.AsObject;
  static serializeBinaryToWriter(message: FactoryList, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryList;
  static deserializeBinaryFromReader(message: FactoryList, reader: jspb.BinaryReader): FactoryList;
}

export namespace FactoryList {
  export type AsObject = {
    factoriesList: Array<FactoryEntry.AsObject>,
  }
}

export class FactoryEntry extends jspb.Message {
  getFactoryId(): FactoryID | undefined;
  setFactoryId(value?: FactoryID): FactoryEntry;
  hasFactoryId(): boolean;
  clearFactoryId(): FactoryEntry;

  getStatus(): FactoryStatus | undefined;
  setStatus(value?: FactoryStatus): FactoryEntry;
  hasStatus(): boolean;
  clearStatus(): FactoryEntry;

  getFactoryType(): string;
  setFactoryType(value: string): FactoryEntry;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FactoryEntry.AsObject;
  static toObject(includeInstance: boolean, msg: FactoryEntry): FactoryEntry.AsObject;
  static serializeBinaryToWriter(message: FactoryEntry, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FactoryEntry;
  static deserializeBinaryFromReader(message: FactoryEntry, reader: jspb.BinaryReader): FactoryEntry;
}

export namespace FactoryEntry {
  export type AsObject = {
    factoryId?: FactoryID.AsObject,
    status?: FactoryStatus.AsObject,
    factoryType: string,
  }
}

export class ActionProposal extends jspb.Message {
  getId(): string;
  setId(value: string): ActionProposal;

  getAgentId(): string;
  setAgentId(value: string): ActionProposal;

  getDescription(): string;
  setDescription(value: string): ActionProposal;

  getPayload(): Uint8Array | string;
  getPayload_asU8(): Uint8Array;
  getPayload_asB64(): string;
  setPayload(value: Uint8Array | string): ActionProposal;

  getPayloadHash(): string;
  setPayloadHash(value: string): ActionProposal;

  getTargetPath(): string;
  setTargetPath(value: string): ActionProposal;

  getIsSecurityAdjacent(): boolean;
  setIsSecurityAdjacent(value: boolean): ActionProposal;

  getSubmittedAtMs(): number;
  setSubmittedAtMs(value: number): ActionProposal;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ActionProposal.AsObject;
  static toObject(includeInstance: boolean, msg: ActionProposal): ActionProposal.AsObject;
  static serializeBinaryToWriter(message: ActionProposal, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ActionProposal;
  static deserializeBinaryFromReader(message: ActionProposal, reader: jspb.BinaryReader): ActionProposal;
}

export namespace ActionProposal {
  export type AsObject = {
    id: string,
    agentId: string,
    description: string,
    payload: Uint8Array | string,
    payloadHash: string,
    targetPath: string,
    isSecurityAdjacent: boolean,
    submittedAtMs: number,
  }
}

export class VerificationEvent extends jspb.Message {
  getProposalId(): string;
  setProposalId(value: string): VerificationEvent;

  getEventType(): VerificationEventType;
  setEventType(value: VerificationEventType): VerificationEvent;

  getTimestampMs(): number;
  setTimestampMs(value: number): VerificationEvent;

  getSafeResult(): SafeResult | undefined;
  setSafeResult(value?: SafeResult): VerificationEvent;
  hasSafeResult(): boolean;
  clearSafeResult(): VerificationEvent;

  getUnsafeResult(): UnsafeResult | undefined;
  setUnsafeResult(value?: UnsafeResult): VerificationEvent;
  hasUnsafeResult(): boolean;
  clearUnsafeResult(): VerificationEvent;

  getErrorResult(): ErrorResult | undefined;
  setErrorResult(value?: ErrorResult): VerificationEvent;
  hasErrorResult(): boolean;
  clearErrorResult(): VerificationEvent;

  getPolicyFingerprintHex(): string;
  setPolicyFingerprintHex(value: string): VerificationEvent;

  getResultCase(): VerificationEvent.ResultCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): VerificationEvent.AsObject;
  static toObject(includeInstance: boolean, msg: VerificationEvent): VerificationEvent.AsObject;
  static serializeBinaryToWriter(message: VerificationEvent, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): VerificationEvent;
  static deserializeBinaryFromReader(message: VerificationEvent, reader: jspb.BinaryReader): VerificationEvent;
}

export namespace VerificationEvent {
  export type AsObject = {
    proposalId: string,
    eventType: VerificationEventType,
    timestampMs: number,
    safeResult?: SafeResult.AsObject,
    unsafeResult?: UnsafeResult.AsObject,
    errorResult?: ErrorResult.AsObject,
    policyFingerprintHex: string,
  }

  export enum ResultCase { 
    RESULT_NOT_SET = 0,
    SAFE_RESULT = 4,
    UNSAFE_RESULT = 5,
    ERROR_RESULT = 6,
  }
}

export class SafeResult extends jspb.Message {
  getProofTier(): SafetyTierProto;
  setProofTier(value: SafetyTierProto): SafeResult;

  getProofCertificateHex(): string;
  setProofCertificateHex(value: string): SafeResult;

  getDurationMs(): number;
  setDurationMs(value: number): SafeResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SafeResult.AsObject;
  static toObject(includeInstance: boolean, msg: SafeResult): SafeResult.AsObject;
  static serializeBinaryToWriter(message: SafeResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SafeResult;
  static deserializeBinaryFromReader(message: SafeResult, reader: jspb.BinaryReader): SafeResult;
}

export namespace SafeResult {
  export type AsObject = {
    proofTier: SafetyTierProto,
    proofCertificateHex: string,
    durationMs: number,
  }
}

export class UnsafeResult extends jspb.Message {
  getViolatedConstraintIdsList(): Array<string>;
  setViolatedConstraintIdsList(value: Array<string>): UnsafeResult;
  clearViolatedConstraintIdsList(): UnsafeResult;
  addViolatedConstraintIds(value: string, index?: number): UnsafeResult;

  getViolationDetailsMap(): jspb.Map<string, string>;
  clearViolationDetailsMap(): UnsafeResult;

  getUnsatCoreIdsList(): Array<string>;
  setUnsatCoreIdsList(value: Array<string>): UnsafeResult;
  clearUnsatCoreIdsList(): UnsafeResult;
  addUnsatCoreIds(value: string, index?: number): UnsafeResult;

  getRemediationHint(): string;
  setRemediationHint(value: string): UnsafeResult;

  getDurationMs(): number;
  setDurationMs(value: number): UnsafeResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnsafeResult.AsObject;
  static toObject(includeInstance: boolean, msg: UnsafeResult): UnsafeResult.AsObject;
  static serializeBinaryToWriter(message: UnsafeResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnsafeResult;
  static deserializeBinaryFromReader(message: UnsafeResult, reader: jspb.BinaryReader): UnsafeResult;
}

export namespace UnsafeResult {
  export type AsObject = {
    violatedConstraintIdsList: Array<string>,
    violationDetailsMap: Array<[string, string]>,
    unsatCoreIdsList: Array<string>,
    remediationHint: string,
    durationMs: number,
  }
}

export class ErrorResult extends jspb.Message {
  getErrorCode(): string;
  setErrorCode(value: string): ErrorResult;

  getMessage(): string;
  setMessage(value: string): ErrorResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ErrorResult.AsObject;
  static toObject(includeInstance: boolean, msg: ErrorResult): ErrorResult.AsObject;
  static serializeBinaryToWriter(message: ErrorResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ErrorResult;
  static deserializeBinaryFromReader(message: ErrorResult, reader: jspb.BinaryReader): ErrorResult;
}

export namespace ErrorResult {
  export type AsObject = {
    errorCode: string,
    message: string,
  }
}

export class ApprovalRequest extends jspb.Message {
  getProposalId(): string;
  setProposalId(value: string): ApprovalRequest;

  getApprovalSignature(): string;
  setApprovalSignature(value: string): ApprovalRequest;

  getApprovedBy(): string;
  setApprovedBy(value: string): ApprovalRequest;

  getApprovalTimestampMs(): number;
  setApprovalTimestampMs(value: number): ApprovalRequest;

  getNote(): string;
  setNote(value: string): ApprovalRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ApprovalRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ApprovalRequest): ApprovalRequest.AsObject;
  static serializeBinaryToWriter(message: ApprovalRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ApprovalRequest;
  static deserializeBinaryFromReader(message: ApprovalRequest, reader: jspb.BinaryReader): ApprovalRequest;
}

export namespace ApprovalRequest {
  export type AsObject = {
    proposalId: string,
    approvalSignature: string,
    approvedBy: string,
    approvalTimestampMs: number,
    note: string,
  }
}

export class MerkleProof extends jspb.Message {
  getLeafHashHex(): string;
  setLeafHashHex(value: string): MerkleProof;

  getTreeSize(): number;
  setTreeSize(value: number): MerkleProof;

  getRootHashHex(): string;
  setRootHashHex(value: string): MerkleProof;

  getInclusionPathList(): Array<string>;
  setInclusionPathList(value: Array<string>): MerkleProof;
  clearInclusionPathList(): MerkleProof;
  addInclusionPath(value: string, index?: number): MerkleProof;

  getSthSignatureHex(): string;
  setSthSignatureHex(value: string): MerkleProof;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MerkleProof.AsObject;
  static toObject(includeInstance: boolean, msg: MerkleProof): MerkleProof.AsObject;
  static serializeBinaryToWriter(message: MerkleProof, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MerkleProof;
  static deserializeBinaryFromReader(message: MerkleProof, reader: jspb.BinaryReader): MerkleProof;
}

export namespace MerkleProof {
  export type AsObject = {
    leafHashHex: string,
    treeSize: number,
    rootHashHex: string,
    inclusionPathList: Array<string>,
    sthSignatureHex: string,
  }
}

export class VetoRequest extends jspb.Message {
  getProposalId(): string;
  setProposalId(value: string): VetoRequest;

  getVetoedBy(): string;
  setVetoedBy(value: string): VetoRequest;

  getRationale(): string;
  setRationale(value: string): VetoRequest;

  getRequiredChangesList(): Array<string>;
  setRequiredChangesList(value: Array<string>): VetoRequest;
  clearRequiredChangesList(): VetoRequest;
  addRequiredChanges(value: string, index?: number): VetoRequest;

  getVetoTimestampMs(): number;
  setVetoTimestampMs(value: number): VetoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): VetoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: VetoRequest): VetoRequest.AsObject;
  static serializeBinaryToWriter(message: VetoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): VetoRequest;
  static deserializeBinaryFromReader(message: VetoRequest, reader: jspb.BinaryReader): VetoRequest;
}

export namespace VetoRequest {
  export type AsObject = {
    proposalId: string,
    vetoedBy: string,
    rationale: string,
    requiredChangesList: Array<string>,
    vetoTimestampMs: number,
  }
}

export class FitnessSnapshot extends jspb.Message {
  getSchemaVersion(): string;
  setSchemaVersion(value: string): FitnessSnapshot;

  getTimestampMs(): number;
  setTimestampMs(value: number): FitnessSnapshot;

  getMetrics(): FitnessMetrics | undefined;
  setMetrics(value?: FitnessMetrics): FitnessSnapshot;
  hasMetrics(): boolean;
  clearMetrics(): FitnessSnapshot;

  getDomainExtensionsMap(): jspb.Map<string, DomainMetricValue>;
  clearDomainExtensionsMap(): FitnessSnapshot;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FitnessSnapshot.AsObject;
  static toObject(includeInstance: boolean, msg: FitnessSnapshot): FitnessSnapshot.AsObject;
  static serializeBinaryToWriter(message: FitnessSnapshot, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FitnessSnapshot;
  static deserializeBinaryFromReader(message: FitnessSnapshot, reader: jspb.BinaryReader): FitnessSnapshot;
}

export namespace FitnessSnapshot {
  export type AsObject = {
    schemaVersion: string,
    timestampMs: number,
    metrics?: FitnessMetrics.AsObject,
    domainExtensionsMap: Array<[string, DomainMetricValue.AsObject]>,
  }
}

export class FitnessMetrics extends jspb.Message {
  getSafetyCompliance(): MetricValue | undefined;
  setSafetyCompliance(value?: MetricValue): FitnessMetrics;
  hasSafetyCompliance(): boolean;
  clearSafetyCompliance(): FitnessMetrics;

  getAuditIntegrity(): MetricValue | undefined;
  setAuditIntegrity(value?: MetricValue): FitnessMetrics;
  hasAuditIntegrity(): boolean;
  clearAuditIntegrity(): FitnessMetrics;

  getDhammaAlignment(): MetricValue | undefined;
  setDhammaAlignment(value?: MetricValue): FitnessMetrics;
  hasDhammaAlignment(): boolean;
  clearDhammaAlignment(): FitnessMetrics;

  getSystemPerformance(): CompositeMetricValue | undefined;
  setSystemPerformance(value?: CompositeMetricValue): FitnessMetrics;
  hasSystemPerformance(): boolean;
  clearSystemPerformance(): FitnessMetrics;

  getOperationalCost(): MetricValue | undefined;
  setOperationalCost(value?: MetricValue): FitnessMetrics;
  hasOperationalCost(): boolean;
  clearOperationalCost(): FitnessMetrics;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FitnessMetrics.AsObject;
  static toObject(includeInstance: boolean, msg: FitnessMetrics): FitnessMetrics.AsObject;
  static serializeBinaryToWriter(message: FitnessMetrics, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FitnessMetrics;
  static deserializeBinaryFromReader(message: FitnessMetrics, reader: jspb.BinaryReader): FitnessMetrics;
}

export namespace FitnessMetrics {
  export type AsObject = {
    safetyCompliance?: MetricValue.AsObject,
    auditIntegrity?: MetricValue.AsObject,
    dhammaAlignment?: MetricValue.AsObject,
    systemPerformance?: CompositeMetricValue.AsObject,
    operationalCost?: MetricValue.AsObject,
  }
}

export class MetricValue extends jspb.Message {
  getCurrentValue(): number;
  setCurrentValue(value: number): MetricValue;

  getPreviousValue(): number;
  setPreviousValue(value: number): MetricValue;

  getStatus(): MetricStatus;
  setStatus(value: MetricStatus): MetricValue;

  getLastUpdatedMs(): number;
  setLastUpdatedMs(value: number): MetricValue;

  getEscalationTriggered(): boolean;
  setEscalationTriggered(value: boolean): MetricValue;

  getNote(): string;
  setNote(value: string): MetricValue;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MetricValue.AsObject;
  static toObject(includeInstance: boolean, msg: MetricValue): MetricValue.AsObject;
  static serializeBinaryToWriter(message: MetricValue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MetricValue;
  static deserializeBinaryFromReader(message: MetricValue, reader: jspb.BinaryReader): MetricValue;
}

export namespace MetricValue {
  export type AsObject = {
    currentValue: number,
    previousValue: number,
    status: MetricStatus,
    lastUpdatedMs: number,
    escalationTriggered: boolean,
    note: string,
  }
}

export class CompositeMetricValue extends jspb.Message {
  getSubMetricsMap(): jspb.Map<string, number>;
  clearSubMetricsMap(): CompositeMetricValue;

  getStatus(): MetricStatus;
  setStatus(value: MetricStatus): CompositeMetricValue;

  getLastUpdatedMs(): number;
  setLastUpdatedMs(value: number): CompositeMetricValue;

  getEscalationTriggered(): boolean;
  setEscalationTriggered(value: boolean): CompositeMetricValue;

  getNote(): string;
  setNote(value: string): CompositeMetricValue;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CompositeMetricValue.AsObject;
  static toObject(includeInstance: boolean, msg: CompositeMetricValue): CompositeMetricValue.AsObject;
  static serializeBinaryToWriter(message: CompositeMetricValue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CompositeMetricValue;
  static deserializeBinaryFromReader(message: CompositeMetricValue, reader: jspb.BinaryReader): CompositeMetricValue;
}

export namespace CompositeMetricValue {
  export type AsObject = {
    subMetricsMap: Array<[string, number]>,
    status: MetricStatus,
    lastUpdatedMs: number,
    escalationTriggered: boolean,
    note: string,
  }
}

export class DomainMetricValue extends jspb.Message {
  getMetricName(): string;
  setMetricName(value: string): DomainMetricValue;

  getValue(): number;
  setValue(value: number): DomainMetricValue;

  getUnit(): string;
  setUnit(value: string): DomainMetricValue;

  getStatus(): MetricStatus;
  setStatus(value: MetricStatus): DomainMetricValue;

  getLastUpdatedMs(): number;
  setLastUpdatedMs(value: number): DomainMetricValue;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DomainMetricValue.AsObject;
  static toObject(includeInstance: boolean, msg: DomainMetricValue): DomainMetricValue.AsObject;
  static serializeBinaryToWriter(message: DomainMetricValue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DomainMetricValue;
  static deserializeBinaryFromReader(message: DomainMetricValue, reader: jspb.BinaryReader): DomainMetricValue;
}

export namespace DomainMetricValue {
  export type AsObject = {
    metricName: string,
    value: number,
    unit: string,
    status: MetricStatus,
    lastUpdatedMs: number,
  }
}

export class DomainFitnessExtension extends jspb.Message {
  getFactoryId(): string;
  setFactoryId(value: string): DomainFitnessExtension;

  getFactoryType(): string;
  setFactoryType(value: string): DomainFitnessExtension;

  getMetricsList(): Array<DomainMetricDeclaration>;
  setMetricsList(value: Array<DomainMetricDeclaration>): DomainFitnessExtension;
  clearMetricsList(): DomainFitnessExtension;
  addMetrics(value?: DomainMetricDeclaration, index?: number): DomainMetricDeclaration;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DomainFitnessExtension.AsObject;
  static toObject(includeInstance: boolean, msg: DomainFitnessExtension): DomainFitnessExtension.AsObject;
  static serializeBinaryToWriter(message: DomainFitnessExtension, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DomainFitnessExtension;
  static deserializeBinaryFromReader(message: DomainFitnessExtension, reader: jspb.BinaryReader): DomainFitnessExtension;
}

export namespace DomainFitnessExtension {
  export type AsObject = {
    factoryId: string,
    factoryType: string,
    metricsList: Array<DomainMetricDeclaration.AsObject>,
  }
}

export class DomainMetricDeclaration extends jspb.Message {
  getMetricId(): string;
  setMetricId(value: string): DomainMetricDeclaration;

  getDisplayName(): string;
  setDisplayName(value: string): DomainMetricDeclaration;

  getDescription(): string;
  setDescription(value: string): DomainMetricDeclaration;

  getUnit(): string;
  setUnit(value: string): DomainMetricDeclaration;

  getDirection(): MetricDirection;
  setDirection(value: MetricDirection): DomainMetricDeclaration;

  getEscalationThreshold(): number;
  setEscalationThreshold(value: number): DomainMetricDeclaration;

  getEscalationOperator(): ThresholdOperator;
  setEscalationOperator(value: ThresholdOperator): DomainMetricDeclaration;

  getOtelMetricName(): string;
  setOtelMetricName(value: string): DomainMetricDeclaration;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DomainMetricDeclaration.AsObject;
  static toObject(includeInstance: boolean, msg: DomainMetricDeclaration): DomainMetricDeclaration.AsObject;
  static serializeBinaryToWriter(message: DomainMetricDeclaration, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DomainMetricDeclaration;
  static deserializeBinaryFromReader(message: DomainMetricDeclaration, reader: jspb.BinaryReader): DomainMetricDeclaration;
}

export namespace DomainMetricDeclaration {
  export type AsObject = {
    metricId: string,
    displayName: string,
    description: string,
    unit: string,
    direction: MetricDirection,
    escalationThreshold: number,
    escalationOperator: ThresholdOperator,
    otelMetricName: string,
  }
}

export class DataContext extends jspb.Message {
  getProposalId(): string;
  setProposalId(value: string): DataContext;

  getDescription(): string;
  setDescription(value: string): DataContext;

  getDomainContextJson(): Uint8Array | string;
  getDomainContextJson_asU8(): Uint8Array;
  getDomainContextJson_asB64(): string;
  setDomainContextJson(value: Uint8Array | string): DataContext;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DataContext.AsObject;
  static toObject(includeInstance: boolean, msg: DataContext): DataContext.AsObject;
  static serializeBinaryToWriter(message: DataContext, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DataContext;
  static deserializeBinaryFromReader(message: DataContext, reader: jspb.BinaryReader): DataContext;
}

export namespace DataContext {
  export type AsObject = {
    proposalId: string,
    description: string,
    domainContextJson: Uint8Array | string,
  }
}

export class ReflectionResult extends jspb.Message {
  getProposalId(): string;
  setProposalId(value: string): ReflectionResult;

  getMoralWeighting(): MoralWeighting | undefined;
  setMoralWeighting(value?: MoralWeighting): ReflectionResult;
  hasMoralWeighting(): boolean;
  clearMoralWeighting(): ReflectionResult;

  getRetrievedSegmentsList(): Array<BIlaraSegment>;
  setRetrievedSegmentsList(value: Array<BIlaraSegment>): ReflectionResult;
  clearRetrievedSegmentsList(): ReflectionResult;
  addRetrievedSegments(value?: BIlaraSegment, index?: number): BIlaraSegment;

  getComputedAtMs(): number;
  setComputedAtMs(value: number): ReflectionResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReflectionResult.AsObject;
  static toObject(includeInstance: boolean, msg: ReflectionResult): ReflectionResult.AsObject;
  static serializeBinaryToWriter(message: ReflectionResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReflectionResult;
  static deserializeBinaryFromReader(message: ReflectionResult, reader: jspb.BinaryReader): ReflectionResult;
}

export namespace ReflectionResult {
  export type AsObject = {
    proposalId: string,
    moralWeighting?: MoralWeighting.AsObject,
    retrievedSegmentsList: Array<BIlaraSegment.AsObject>,
    computedAtMs: number,
  }
}

export class MoralWeighting extends jspb.Message {
  getScore(): number;
  setScore(value: number): MoralWeighting;

  getRoot(): MoralRoot;
  setRoot(value: MoralRoot): MoralWeighting;

  getCitationIdsList(): Array<string>;
  setCitationIdsList(value: Array<string>): MoralWeighting;
  clearCitationIdsList(): MoralWeighting;
  addCitationIds(value: string, index?: number): MoralWeighting;

  getReasoning(): string;
  setReasoning(value: string): MoralWeighting;

  getSemanticMapVersion(): string;
  setSemanticMapVersion(value: string): MoralWeighting;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MoralWeighting.AsObject;
  static toObject(includeInstance: boolean, msg: MoralWeighting): MoralWeighting.AsObject;
  static serializeBinaryToWriter(message: MoralWeighting, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MoralWeighting;
  static deserializeBinaryFromReader(message: MoralWeighting, reader: jspb.BinaryReader): MoralWeighting;
}

export namespace MoralWeighting {
  export type AsObject = {
    score: number,
    root: MoralRoot,
    citationIdsList: Array<string>,
    reasoning: string,
    semanticMapVersion: string,
  }
}

export class BIlaraSegment extends jspb.Message {
  getSegmentId(): string;
  setSegmentId(value: string): BIlaraSegment;

  getPaliText(): string;
  setPaliText(value: string): BIlaraSegment;

  getTranslation(): string;
  setTranslation(value: string): BIlaraSegment;

  getStylometricScore(): number;
  setStylometricScore(value: number): BIlaraSegment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BIlaraSegment.AsObject;
  static toObject(includeInstance: boolean, msg: BIlaraSegment): BIlaraSegment.AsObject;
  static serializeBinaryToWriter(message: BIlaraSegment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BIlaraSegment;
  static deserializeBinaryFromReader(message: BIlaraSegment, reader: jspb.BinaryReader): BIlaraSegment;
}

export namespace BIlaraSegment {
  export type AsObject = {
    segmentId: string,
    paliText: string,
    translation: string,
    stylometricScore: number,
  }
}

export class Briefing extends jspb.Message {
  getBriefingId(): string;
  setBriefingId(value: string): Briefing;

  getTopic(): string;
  setTopic(value: string): Briefing;

  getAuthor(): string;
  setAuthor(value: string): Briefing;

  getPhase(): string;
  setPhase(value: string): Briefing;

  getSummaryMarkdown(): string;
  setSummaryMarkdown(value: string): Briefing;

  getArtifactsList(): Array<string>;
  setArtifactsList(value: Array<string>): Briefing;
  clearArtifactsList(): Briefing;
  addArtifacts(value: string, index?: number): Briefing;

  getQuestionsList(): Array<string>;
  setQuestionsList(value: Array<string>): Briefing;
  clearQuestionsList(): Briefing;
  addQuestions(value: string, index?: number): Briefing;

  getSubmittedAtMs(): number;
  setSubmittedAtMs(value: number): Briefing;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Briefing.AsObject;
  static toObject(includeInstance: boolean, msg: Briefing): Briefing.AsObject;
  static serializeBinaryToWriter(message: Briefing, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Briefing;
  static deserializeBinaryFromReader(message: Briefing, reader: jspb.BinaryReader): Briefing;
}

export namespace Briefing {
  export type AsObject = {
    briefingId: string,
    topic: string,
    author: string,
    phase: string,
    summaryMarkdown: string,
    artifactsList: Array<string>,
    questionsList: Array<string>,
    submittedAtMs: number,
  }
}

export class VerdictQuery extends jspb.Message {
  getTopic(): string;
  setTopic(value: string): VerdictQuery;

  getProposalId(): string;
  setProposalId(value: string): VerdictQuery;

  getBriefingId(): string;
  setBriefingId(value: string): VerdictQuery;

  getQueryCase(): VerdictQuery.QueryCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): VerdictQuery.AsObject;
  static toObject(includeInstance: boolean, msg: VerdictQuery): VerdictQuery.AsObject;
  static serializeBinaryToWriter(message: VerdictQuery, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): VerdictQuery;
  static deserializeBinaryFromReader(message: VerdictQuery, reader: jspb.BinaryReader): VerdictQuery;
}

export namespace VerdictQuery {
  export type AsObject = {
    topic: string,
    proposalId: string,
    briefingId: string,
  }

  export enum QueryCase { 
    QUERY_NOT_SET = 0,
    TOPIC = 1,
    PROPOSAL_ID = 2,
    BRIEFING_ID = 3,
  }
}

export class AnalystVerdict extends jspb.Message {
  getVerdictId(): string;
  setVerdictId(value: string): AnalystVerdict;

  getArtifactPath(): string;
  setArtifactPath(value: string): AnalystVerdict;

  getVerdict(): VerdictDecision;
  setVerdict(value: VerdictDecision): AnalystVerdict;

  getRationale(): string;
  setRationale(value: string): AnalystVerdict;

  getRequiredChangesList(): Array<string>;
  setRequiredChangesList(value: Array<string>): AnalystVerdict;
  clearRequiredChangesList(): AnalystVerdict;
  addRequiredChanges(value: string, index?: number): AnalystVerdict;

  getFitnessImpactNotes(): string;
  setFitnessImpactNotes(value: string): AnalystVerdict;

  getIssuedAtMs(): number;
  setIssuedAtMs(value: number): AnalystVerdict;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AnalystVerdict.AsObject;
  static toObject(includeInstance: boolean, msg: AnalystVerdict): AnalystVerdict.AsObject;
  static serializeBinaryToWriter(message: AnalystVerdict, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AnalystVerdict;
  static deserializeBinaryFromReader(message: AnalystVerdict, reader: jspb.BinaryReader): AnalystVerdict;
}

export namespace AnalystVerdict {
  export type AsObject = {
    verdictId: string,
    artifactPath: string,
    verdict: VerdictDecision,
    rationale: string,
    requiredChangesList: Array<string>,
    fitnessImpactNotes: string,
    issuedAtMs: number,
  }
}

export class RegistrationResult extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): RegistrationResult;

  getFactoryId(): string;
  setFactoryId(value: string): RegistrationResult;

  getRegisteredCount(): number;
  setRegisteredCount(value: number): RegistrationResult;

  getErrorMessage(): string;
  setErrorMessage(value: string): RegistrationResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegistrationResult.AsObject;
  static toObject(includeInstance: boolean, msg: RegistrationResult): RegistrationResult.AsObject;
  static serializeBinaryToWriter(message: RegistrationResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegistrationResult;
  static deserializeBinaryFromReader(message: RegistrationResult, reader: jspb.BinaryReader): RegistrationResult;
}

export namespace RegistrationResult {
  export type AsObject = {
    success: boolean,
    factoryId: string,
    registeredCount: number,
    errorMessage: string,
  }
}

export enum FactoryState { 
  FACTORY_STATE_UNSPECIFIED = 0,
  FACTORY_STARTING = 1,
  FACTORY_RUNNING = 2,
  FACTORY_STOPPING = 3,
  FACTORY_STOPPED = 4,
  FACTORY_ERROR = 5,
}
export enum VerificationEventType { 
  VERIFICATION_EVENT_UNSPECIFIED = 0,
  VERIFICATION_RECEIVED = 1,
  VERIFICATION_CHECKING = 2,
  VERIFICATION_SAFE = 3,
  VERIFICATION_UNSAFE = 4,
  VERIFICATION_GATE_PENDING = 5,
  VERIFICATION_GATE_APPROVED = 6,
  VERIFICATION_GATE_DENIED = 7,
  VERIFICATION_COMMITTED = 8,
  VERIFICATION_VETOED = 9,
  VERIFICATION_TIMEOUT = 10,
  VERIFICATION_ERROR = 11,
}
export enum SafetyTierProto { 
  SAFETY_TIER_UNSPECIFIED = 0,
  SAFETY_TIER_1 = 1,
  SAFETY_TIER_2 = 2,
}
export enum MetricStatus { 
  METRIC_STATUS_UNSPECIFIED = 0,
  METRIC_UNINITIALIZED = 1,
  METRIC_GREEN = 2,
  METRIC_AMBER = 3,
  METRIC_RED = 4,
}
export enum MetricDirection { 
  METRIC_DIRECTION_UNSPECIFIED = 0,
  METRIC_LOWER_IS_BETTER = 1,
  METRIC_HIGHER_IS_BETTER = 2,
}
export enum ThresholdOperator { 
  THRESHOLD_OPERATOR_UNSPECIFIED = 0,
  THRESHOLD_GTE = 1,
  THRESHOLD_GT = 2,
  THRESHOLD_LTE = 3,
  THRESHOLD_LT = 4,
}
export enum MoralRoot { 
  MORAL_ROOT_UNSPECIFIED = 0,
  MORAL_ROOT_KUSALA = 1,
  MORAL_ROOT_AKUSALA = 2,
  MORAL_ROOT_NEUTRAL = 3,
}
export enum VerdictDecision { 
  VERDICT_UNSPECIFIED = 0,
  VERDICT_APPROVED = 1,
  VERDICT_VETOED = 2,
  VERDICT_CONDITIONAL = 3,
}
