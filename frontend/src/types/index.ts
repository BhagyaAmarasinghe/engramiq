export interface Site {
  id: string;
  site_code: string;
  name: string;
  address: string;
  country: string;
  total_capacity_kw: number;
  number_of_inverters: number;
  installation_date: string;
  site_metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface Component {
  id: string;
  siteId: string;
  externalId?: string;
  componentType: ComponentType;
  name: string;
  label?: string;
  level: number;
  groupName?: string;
  specifications: Record<string, any>;
  electricalData: Record<string, any>;
  physicalData: Record<string, any>;
  drawingTitle?: string;
  drawingNumber?: string;
  revision?: string;
  revisionDate?: string;
  spatialId?: string;
  coordinates?: string;
  currentStatus: ComponentStatus;
  lastMaintenanceDate?: string;
  nextMaintenanceDate?: string;
  createdAt: string;
  updatedAt: string;
}

export type ComponentType = 
  | 'inverter' 
  | 'combiner' 
  | 'panel' 
  | 'transformer' 
  | 'meter' 
  | 'switchgear' 
  | 'monitoring' 
  | 'other';

export type ComponentStatus = 
  | 'operational' 
  | 'fault' 
  | 'maintenance' 
  | 'offline';

export interface Document {
  id: string;
  siteId: string;
  title: string;
  documentType: DocumentType;
  filePath?: string;
  fileSize?: number;
  fileHash?: string;
  mimeType?: string;
  rawContent?: string;
  processedContent?: string;
  processingStatus: ProcessingStatus;
  processingError?: string;
  processingStartedAt?: string;
  processingCompletedAt?: string;
  sourceType?: string;
  sourceIdentifier?: string;
  originalFilename?: string;
  authorName?: string;
  authorEmail?: string;
  uploadedBy?: string;
  documentMetadata: Record<string, any>;
  uploadUserId?: string;
  createdAt: string;
  updatedAt: string;
}

export type DocumentType = 
  | 'field_service_report'
  | 'email' 
  | 'meeting_transcript'
  | 'work_order' 
  | 'inspection_report' 
  | 'warranty_claim'
  | 'contract' 
  | 'manual' 
  | 'drawing' 
  | 'other';

export type ProcessingStatus = 
  | 'pending' 
  | 'processing' 
  | 'completed' 
  | 'failed';

export interface UserQuery {
  id: string;
  siteId?: string;
  userId?: string;
  queryText: string;
  queryType?: string;
  enhanced: boolean;
  answer?: string;
  confidenceScore?: number;
  extractedEntities: Record<string, any>;
  relatedConcepts: string[];
  responseType?: string;
  noHallucination: boolean;
  processingTimeMs?: number;
  createdAt: string;
}

export interface QueryResponse {
  answer: string;
  confidence: number;
  sources: SourceAttribution[];
  processingTime: number;
  noHallucination: boolean;
  responseType: string;
}

export interface SourceAttribution {
  documentId: string;
  documentTitle: string;
  relevantExcerpt: string;
  pageNumber?: number;
  confidence: number;
  citation: string;
}

export interface ExtractedAction {
  id: string;
  documentId: string;
  siteId: string;
  actionType: ActionType;
  description: string;
  actionDate?: string;
  status: ActionStatus;
  componentType?: string;
  componentNames: string[];
  technicianNames: string[];
  workOrderNumber?: string;
  durationHours?: number;
  measurements: Record<string, any>;
  partsUsed: Record<string, any>;
  confidenceScore?: number;
  createdAt: string;
  updatedAt: string;
}

export type ActionType = 
  | 'maintenance'
  | 'replacement'
  | 'troubleshoot'
  | 'inspection'
  | 'repair'
  | 'testing'
  | 'installation'
  | 'commissioning'
  | 'fault_clearing'
  | 'monitoring'
  | 'cleaning'
  | 'other';

export type ActionStatus = 
  | 'planned'
  | 'in_progress'
  | 'completed'
  | 'cancelled'
  | 'on_hold'
  | 'requires_follow_up';

export interface TimelineEvent {
  id: string;
  siteId: string;
  eventType: EventType;
  title: string;
  description?: string;
  priority: EventPriority;
  startTime: string;
  endTime?: string;
  durationMinutes?: number;
  affectedComponentIds: string[];
  relatedDocumentId?: string;
  relatedActionId?: string;
  assignedTechnicians: string[];
  responsibleUserId?: string;
  eventMetadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export type EventType = 
  | 'maintenance_scheduled'
  | 'maintenance_completed'
  | 'fault_occurred'
  | 'fault_cleared'
  | 'replacement_scheduled'
  | 'replacement_completed'
  | 'inspection_scheduled'
  | 'inspection_completed'
  | 'warranty_claim'
  | 'performance_alert'
  | 'contract_milestone'
  | 'other';

export type EventPriority = 'low' | 'medium' | 'high' | 'critical';