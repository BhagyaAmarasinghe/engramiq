import axios from 'axios';
import { 
  Site, 
  Component, 
  Document, 
  UserQuery, 
  QueryResponse, 
  ExtractedAction, 
  TimelineEvent 
} from '@/types';
import { toCamelCase } from '@/lib/utils';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
});

// Request interceptor for auth tokens and content type
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // Set Content-Type to application/json only if it's not FormData
  if (!(config.data instanceof FormData)) {
    config.headers['Content-Type'] = 'application/json';
  }

  return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

// Site APIs
export const siteAPI = {
  getSites: async (): Promise<{ data: Site[] }> => {
    const response = await api.get('/sites');
    return response.data; // Return the actual data, not the axios response wrapper
  },
  
  getSite: async (siteId: string): Promise<{ data: Site }> => {
    const response = await api.get(`/sites/${siteId}`);
    return response.data;
  },
  
  createSite: async (site: Partial<Site>): Promise<{ data: Site }> => {
    const response = await api.post('/sites', site);
    return response.data;
  },
  
  updateSite: async (siteId: string, site: Partial<Site>): Promise<{ data: Site }> => {
    const response = await api.put(`/sites/${siteId}`, site);
    return response.data;
  },
  
  deleteSite: async (siteId: string): Promise<void> => {
    await api.delete(`/sites/${siteId}`);
  },
};

// Component APIs
export const componentAPI = {
  getComponents: async (siteId: string, params?: Record<string, any>): Promise<{ data: Component[] }> => {
    const response = await api.get(`/sites/${siteId}/components`, { params });
    // Backend returns { components: [...] }, but we need { data: [...] }
    const components = response.data.components || [];

    // Convert snake_case to camelCase for each component
    const transformedComponents = components.map((comp: any) => ({
      id: comp.id,
      siteId: comp.site_id,
      externalId: comp.external_id,
      componentType: comp.component_type,
      name: comp.name,
      label: comp.label,
      level: comp.level,
      groupName: comp.group_name,
      specifications: comp.specifications,
      electricalData: comp.electrical_data,
      physicalData: comp.physical_data,
      drawingTitle: comp.drawing_title,
      drawingNumber: comp.drawing_number,
      revision: comp.revision,
      revisionDate: comp.revision_date,
      spatialId: comp.spatial_id,
      coordinates: comp.coordinates,
      currentStatus: comp.current_status,
      lastMaintenanceDate: comp.last_maintenance_date,
      nextMaintenanceDate: comp.next_maintenance_date,
      createdAt: comp.created_at,
      updatedAt: comp.updated_at
    }));

    return { data: transformedComponents };
  },
  
  getComponent: (componentId: string): Promise<{ data: Component }> => 
    api.get(`/components/${componentId}`),
  
  createComponent: (siteId: string, component: Partial<Component>): Promise<{ data: Component }> => 
    api.post(`/sites/${siteId}/components`, component),
  
  updateComponent: (componentId: string, component: Partial<Component>): Promise<{ data: Component }> => 
    api.put(`/components/${componentId}`, component),
  
  deleteComponent: (componentId: string): Promise<void> => 
    api.delete(`/components/${componentId}`),
  
  bulkCreateComponents: (siteId: string, components: Partial<Component>[]): Promise<{ data: Component[] }> => 
    api.post(`/sites/${siteId}/components/bulk`, { components }),
};

// Document APIs
export const documentAPI = {
  getDocuments: async (siteId: string, params?: Record<string, any>): Promise<{ data: Document[] }> => {
    const response = await api.get(`/sites/${siteId}/documents`, { params });
    // Backend returns { documents: [...] }, but we need { data: [...] }
    const documents = response.data.documents || [];

    // Convert snake_case to camelCase for each document
    const transformedDocuments = documents.map((doc: any) => ({
      id: doc.id,
      siteId: doc.site_id,
      title: doc.title,
      documentType: doc.document_type,
      filePath: doc.file_path,
      fileSize: doc.file_size,
      fileHash: doc.file_hash,
      mimeType: doc.mime_type,
      rawContent: doc.raw_content,
      processedContent: doc.processed_content,
      processingStatus: doc.processing_status,
      processingError: doc.processing_error,
      processingStartedAt: doc.processing_started_at,
      processingCompletedAt: doc.processing_completed_at,
      sourceType: doc.source_type,
      sourceIdentifier: doc.source_identifier,
      originalFilename: doc.original_filename,
      authorName: doc.author_name,
      authorEmail: doc.author_email,
      uploadedBy: doc.uploaded_by,
      documentMetadata: doc.document_metadata,
      uploadUserId: doc.upload_user_id,
      createdAt: doc.created_at,
      updatedAt: doc.updated_at
    }));

    return { data: transformedDocuments };
  },
  
  getDocument: (documentId: string): Promise<{ data: Document }> => 
    api.get(`/documents/${documentId}`),
  
  uploadDocument: (siteId: string, file: File, documentType: string, metadata?: Record<string, any>) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('document_type', documentType);
    if (metadata) {
      Object.entries(metadata).forEach(([key, value]) => {
        formData.append(key, JSON.stringify(value));
      });
    }

    // Don't set Content-Type header, let the browser set it with boundary
    return api.post(`/sites/${siteId}/documents`, formData);
  },
  
  processDocument: (documentId: string): Promise<{ data: Document }> => 
    api.post(`/documents/${documentId}/process`),
  
  deleteDocument: (documentId: string): Promise<void> => 
    api.delete(`/documents/${documentId}`),
  
  searchDocuments: (siteId: string, params: { 
    query: string; 
    semantic?: boolean; 
    documentType?: string;
    limit?: number;
  }): Promise<{ data: Document[] }> => 
    api.get(`/sites/${siteId}/documents/search`, { params }),
};

// Query APIs
export const queryAPI = {
  createQuery: (siteId: string, queryData: {
    query_text: string;
    enhanced?: boolean;
    context?: Record<string, any>;
  }): Promise<{ data: QueryResponse }> => 
    api.post(`/sites/${siteId}/queries`, queryData),
  
  getQuery: (queryId: string): Promise<{ data: UserQuery }> => 
    api.get(`/queries/${queryId}`),
  
  getQueryHistory: (siteId?: string, params?: Record<string, any>): Promise<{ data: UserQuery[] }> => {
    const endpoint = siteId ? `/sites/${siteId}/queries/history` : '/queries/history';
    return api.get(endpoint, { params });
  },
  
  getSimilarQueries: (siteId: string, queryText: string, limit = 5): Promise<{ data: UserQuery[] }> => 
    api.get(`/sites/${siteId}/queries/similar`, { 
      params: { query_text: queryText, limit } 
    }),
};

// Action APIs
export const actionAPI = {
  getActions: (siteId: string, params?: Record<string, any>): Promise<{ data: ExtractedAction[] }> => 
    api.get(`/sites/${siteId}/actions`, { params }),
  
  getAction: (actionId: string): Promise<{ data: ExtractedAction }> => 
    api.get(`/actions/${actionId}`),
  
  getComponentActions: (componentId: string): Promise<{ data: ExtractedAction[] }> => 
    api.get(`/components/${componentId}/actions`),
};

// Timeline APIs
export const timelineAPI = {
  getTimeline: async (siteId: string, params?: {
    startDate?: string;
    endDate?: string;
    eventType?: string;
    priority?: string;
    limit?: number;
  }): Promise<{ data: TimelineEvent[] }> => {
    const response = await api.get(`/sites/${siteId}/timeline`, { params });
    // Backend returns { actions: [...] }, but we need { data: [...] }
    return { data: response.data.actions || [] };
  },
};

// Health check
export const healthAPI = {
  check: (): Promise<{ data: { status: string; timestamp: string } }> => 
    api.get('/health'),
};