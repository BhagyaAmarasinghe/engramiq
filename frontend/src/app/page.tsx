'use client';
// Force client-side only rendering to ensure data loads properly
import dynamic from 'next/dynamic';

import { useState, useEffect } from 'react';
import { Card, CardBody, Button, Spinner, Spacer } from '@heroui/react';
import { IconRefresh, IconAlertCircle } from '@tabler/icons-react';

// Components
import { Sidebar } from '@/components/layout/Sidebar';
import { SiteOverview } from '@/components/dashboard/SiteOverview';
import { QueryInterface } from '@/components/ui/QueryInterface';
import { DocumentUpload } from '@/components/ui/DocumentUpload';
import { Timeline } from '@/components/dashboard/Timeline';

// API and Types
import { siteAPI, componentAPI, documentAPI, queryAPI, timelineAPI } from '@/lib/api';
import { Site, Component, Document, TimelineEvent, QueryResponse } from '@/types';
import { parseErrorMessage } from '@/lib/utils';

export default function HomePage() {
  const [activeTab, setActiveTab] = useState('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Data states
  const [site, setSite] = useState<Site | null>(null);
  const [components, setComponents] = useState<Component[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [timelineEvents, setTimelineEvents] = useState<TimelineEvent[]>([]);
  const [actions, setActions] = useState([]);

  // Default site ID (S2367 from the backend)
  const siteId = 'site-s2367'; // This would come from routing or auth in a real app

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);

      // In a real app, you'd get the site ID from the route or user context
      // For now, we'll load the first site or use a default
      const sitesResponse = await siteAPI.getSites();
      console.log('Sites response structure:', {
        hasData: !!sitesResponse?.data,
        dataIsArray: Array.isArray(sitesResponse?.data),
        dataLength: sitesResponse?.data?.length,
        firstItem: sitesResponse?.data?.[0]
      });
      
      // The response should have a data property containing the sites array
      const sites = sitesResponse?.data;
      
      if (!sites || !Array.isArray(sites) || sites.length === 0) {
        console.error('Invalid sites response:', sitesResponse);
        throw new Error('No sites found. Please check your backend connection.');
      }
      
      const currentSite = sites[0]; // Use first site
      console.log('Current site:', currentSite); // Debug log
      
      // Ensure we have a valid site with an ID before proceeding
      if (!currentSite || !currentSite.id) {
        console.error('Invalid site data:', currentSite);
        throw new Error('Invalid site data received from API');
      }
      
      setSite(currentSite);

      // Load related data in parallel - handle both camelCase and snake_case
      const siteId = currentSite.id || (currentSite as any).id;
      const [componentsRes, documentsRes, timelineRes] = await Promise.allSettled([
        componentAPI.getComponents(siteId),
        documentAPI.getDocuments(siteId),
        timelineAPI.getTimeline(siteId)
      ]);

      // Handle components
      if (componentsRes.status === 'fulfilled') {
        console.log('Components response:', componentsRes.value);
        setComponents(componentsRes.value.data || []);
      } else {
        console.warn('Failed to load components:', componentsRes.reason);
        setComponents([]);
      }

      // Handle documents
      if (documentsRes.status === 'fulfilled') {
        setDocuments(documentsRes.value.data);
      } else {
        console.warn('Failed to load documents:', documentsRes.reason);
        setDocuments([]);
      }

      // Handle timeline
      if (timelineRes.status === 'fulfilled') {
        setTimelineEvents(timelineRes.value.data);
      } else {
        console.warn('Failed to load timeline:', timelineRes.reason);
        setTimelineEvents([]);
      }

    } catch (err: any) {
      console.error('Failed to load data:', err);
      setError(parseErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleQuerySubmit = async (query: string): Promise<QueryResponse> => {
    if (!site) {
      throw new Error('No site selected');
    }

    const response = await queryAPI.createQuery(site.id, {
      query_text: query,
      enhanced: true
    });

    return response.data;
  };

  const handleDocumentUpload = (document: any) => {
    // Refresh documents after upload
    if (site) {
      documentAPI.getDocuments(site.id).then(res => {
        setDocuments(res.data);
      }).catch(console.error);
    }
  };

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <Spinner size="lg" color="primary" />
            <p className="text-default-500 mt-4">Loading site data...</p>
          </div>
        </div>
      );
    }

    if (error) {
      return (
        <div className="flex items-center justify-center h-full">
          <Card className="max-w-md">
            <CardBody className="text-center p-8">
              <IconAlertCircle className="w-12 h-12 text-danger mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-danger mb-2">Error Loading Data</h3>
              <p className="text-default-500 mb-4">{error}</p>
              <Button color="primary" onClick={loadData} startContent={<IconRefresh className="w-4 h-4" />}>
                Try Again
              </Button>
            </CardBody>
          </Card>
        </div>
      );
    }

    if (!site) {
      return (
        <div className="flex items-center justify-center h-full">
          <Card className="max-w-md">
            <CardBody className="text-center p-8">
              <h3 className="text-lg font-semibold mb-2">No Site Data</h3>
              <p className="text-default-500 mb-4">
                No site data found. Make sure your backend is running and has sample data loaded.
              </p>
              <Button color="primary" onClick={loadData} startContent={<IconRefresh className="w-4 h-4" />}>
                Reload
              </Button>
            </CardBody>
          </Card>
        </div>
      );
    }

    switch (activeTab) {
      case 'overview':
        return (
          <SiteOverview
            site={site}
            components={components}
            documents={documents}
            actions={actions}
          />
        );

      case 'query':
        return (
          <div className="h-full flex flex-col">
            <div className="mb-6">
              <h1 className="text-2xl font-bold text-white mb-2">Ask EngramIQ</h1>
              <p className="text-default-400">
                Query your site data using natural language. Get insights about maintenance, 
                components, and site activities.
              </p>
            </div>
            <div className="flex-1">
              <QueryInterface
                siteId={site.id}
                onQuerySubmit={handleQuerySubmit}
              />
            </div>
          </div>
        );

      case 'documents':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Documents</h1>
              <p className="text-default-400">
                Upload field service reports, emails, and meeting transcripts. 
                Documents are automatically processed and made searchable.
              </p>
            </div>
            <DocumentUpload
              siteId={site.id}
              onUploadComplete={handleDocumentUpload}
              onUploadError={(error) => console.error('Upload error:', error)}
            />
            
            {/* Document list would go here */}
            <Card className="glass-effect">
              <CardBody className="p-6">
                <h3 className="text-lg font-semibold mb-4">Recent Documents</h3>
                <div className="space-y-3">
                  {(documents || []).slice(0, 5).map((doc) => (
                    <div key={doc.id} className="flex items-center justify-between p-3 rounded-lg bg-default-100/5">
                      <div>
                        <p className="font-medium text-white">{doc.title}</p>
                        <p className="text-sm text-default-400">
                          {doc.documentType.replace('_', ' ')} • {new Date(doc.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                      <div className="text-sm text-default-400">
                        {doc.processingStatus}
                      </div>
                    </div>
                  ))}
                  {(!documents || documents.length === 0) && (
                    <p className="text-center text-default-500 py-8">
                      No documents uploaded yet. Upload your first document above.
                    </p>
                  )}
                </div>
              </CardBody>
            </Card>
          </div>
        );

      case 'components':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Components</h1>
              <p className="text-default-400">
                Monitor solar equipment status and specifications across your site.
              </p>
            </div>
            
            <Card className="glass-effect">
              <CardBody className="p-6">
                <h3 className="text-lg font-semibold mb-4">Site Components</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {(components || []).map((component) => (
                    <Card key={component.id} className="border border-default-200">
                      <CardBody className="p-4">
                        <div className="flex items-center justify-between mb-2">
                          <h4 className="font-medium">{component.name}</h4>
                          <span className={`px-2 py-1 rounded text-xs font-medium ${
                            component.currentStatus === 'operational' ? 'bg-success/20 text-success' :
                            component.currentStatus === 'fault' ? 'bg-danger/20 text-danger' :
                            component.currentStatus === 'maintenance' ? 'bg-warning/20 text-warning' :
                            'bg-default/20 text-default-500'
                          }`}>
                            {component.currentStatus}
                          </span>
                        </div>
                        <p className="text-sm text-default-500 mb-2">
                          {component.componentType.replace('_', ' ')}
                        </p>
                        {component.specifications?.manufacturer && (
                          <p className="text-sm text-default-400">
                            {component.specifications.manufacturer} {component.specifications.model}
                          </p>
                        )}
                      </CardBody>
                    </Card>
                  ))}
                  {(!components || components.length === 0) && (
                    <div className="col-span-full text-center py-8">
                      <p className="text-default-500">No components found.</p>
                    </div>
                  )}
                </div>
              </CardBody>
            </Card>
          </div>
        );

      case 'timeline':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Timeline</h1>
              <p className="text-default-400">
                Track maintenance activities, events, and site history.
              </p>
            </div>
            <Timeline events={timelineEvents} />
          </div>
        );

      case 'actions':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Actions</h1>
              <p className="text-default-400">
                Maintenance actions and activities extracted from your documents.
              </p>
            </div>
            <Card className="glass-effect">
              <CardBody className="text-center py-12">
                <p className="text-default-500">Actions view coming soon...</p>
              </CardBody>
            </Card>
          </div>
        );

      case 'search':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Search</h1>
              <p className="text-default-400">
                Advanced search across all site documents and data.
              </p>
            </div>
            <Card className="glass-effect">
              <CardBody className="text-center py-12">
                <p className="text-default-500">Advanced search coming soon...</p>
              </CardBody>
            </Card>
          </div>
        );

      case 'settings':
        return (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">Settings</h1>
              <p className="text-default-400">
                Configure your EngramIQ preferences and account settings.
              </p>
            </div>
            <Card className="glass-effect">
              <CardBody className="text-center py-12">
                <p className="text-default-500">Settings panel coming soon...</p>
              </CardBody>
            </Card>
          </div>
        );

      default:
        return (
          <div className="text-center py-12">
            <p className="text-default-500">Select a tab to view content</p>
          </div>
        );
    }
  };

  return (
    <div className="flex h-screen bg-primary-dark-blue">
      <Sidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
      />

      <main className="flex-1 overflow-y-auto ml-8">
        <div className="p-8 h-full max-w-7xl">
          {renderContent()}
        </div>
      </main>
    </div>
  );
}