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

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for auth tokens (when implemented)
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
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
    return response.data;
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
    return response.data;
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
    
    return api.post(`/sites/${siteId}/documents`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
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
    return response.data;
  },
};

// Health check
export const healthAPI = {
  check: (): Promise<{ data: { status: string; timestamp: string } }> => 
    api.get('/health'),
};