'use client';

import { useState, useRef, useEffect } from 'react';
import {
  Card,
  CardHeader,
  CardBody,
  Button,
  Textarea,
  Chip,
  Spinner,
  Divider,
  Link,
  Avatar,
  Tooltip
} from '@heroui/react';
import {
  IconSend,
  IconUser,
  IconRobot,
  IconHistory,
  IconCopy,
  IconExternalLink,
  IconCheck,
  IconAlertTriangle,
  IconClock
} from '@tabler/icons-react';
import { cn, formatDateTime, formatConfidenceScore } from '@/lib/utils';
import { QueryResponse, SourceAttribution } from '@/types';

interface QueryInterfaceProps {
  siteId: string;
  onQuerySubmit?: (query: string) => Promise<QueryResponse>;
  className?: string;
}

interface Message {
  id: string;
  type: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  response?: QueryResponse;
  isLoading?: boolean;
  error?: string;
}

export function QueryInterface({ 
  siteId, 
  onQuerySubmit,
  className 
}: QueryInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [currentQuery, setCurrentQuery] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentQuery.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: currentQuery.trim(),
      timestamp: new Date(),
    };

    const loadingMessage: Message = {
      id: (Date.now() + 1).toString(),
      type: 'assistant',
      content: '',
      timestamp: new Date(),
      isLoading: true,
    };

    setMessages(prev => [...prev, userMessage, loadingMessage]);
    setCurrentQuery('');
    setIsLoading(true);

    try {
      const response = await onQuerySubmit?.(currentQuery.trim());
      
      setMessages(prev => prev.map(msg => 
        msg.id === loadingMessage.id 
          ? {
              ...msg,
              content: response?.answer || 'No response received',
              response,
              isLoading: false,
            }
          : msg
      ));
    } catch (error: any) {
      setMessages(prev => prev.map(msg => 
        msg.id === loadingMessage.id 
          ? {
              ...msg,
              content: '',
              error: error.message || 'An error occurred',
              isLoading: false,
            }
          : msg
      ));
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as any);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // Could add toast notification here
  };

  const suggestedQueries = [
    "What maintenance was performed on inverter INV001 last month?",
    "Show me all fault reports from the past week",
    "Which components require preventive maintenance?",
    "What are the most common issues reported?",
  ];

  return (
    <div className={cn('flex flex-col h-full max-w-4xl mx-auto', className)}>
      {/* Chat Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        {messages.length === 0 && (
          <div className="text-center py-12">
            <div className="p-6 rounded-full bg-primary/10 inline-block mb-6">
              <IconRobot className="w-12 h-12 text-primary" />
            </div>
            <h3 className="text-xl font-semibold mb-2">Ask about your solar site</h3>
            <p className="text-default-500 mb-6 max-w-md mx-auto">
              I can help you query maintenance records, component status, and site activities using natural language.
            </p>
            
            {/* Suggested Queries */}
            <div className="space-y-2">
              <p className="text-sm font-medium text-default-600">Try asking:</p>
              <div className="flex flex-wrap gap-2 justify-center">
                {suggestedQueries.map((query, index) => (
                  <Button
                    key={index}
                    size="sm"
                    variant="flat"
                    color="primary"
                    onClick={() => setCurrentQuery(query)}
                    className="text-xs"
                  >
                    {query}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        )}

        {messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              'flex gap-4',
              message.type === 'user' ? 'justify-end' : 'justify-start'
            )}
          >
            {message.type === 'assistant' && (
              <Avatar
                icon={<IconRobot className="w-5 h-5" />}
                className="bg-primary text-white flex-shrink-0"
                size="sm"
              />
            )}

            <div
              className={cn(
                'max-w-3xl',
                message.type === 'user' ? 'flex justify-end' : ''
              )}
            >
              <Card
                className={cn(
                  message.type === 'user'
                    ? 'bg-primary text-white'
                    : 'glass-effect'
                )}
              >
                <CardBody className="p-4">
                  {message.isLoading ? (
                    <div className="flex items-center gap-3">
                      <Spinner size="sm" color="primary" />
                      <span className="text-default-500">Processing your query...</span>
                    </div>
                  ) : message.error ? (
                    <div className="flex items-start gap-3">
                      <IconAlertTriangle className="w-5 h-5 text-danger flex-shrink-0 mt-0.5" />
                      <div>
                        <p className="text-danger font-medium">Error processing query</p>
                        <p className="text-default-500 text-sm mt-1">{message.error}</p>
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <p className="whitespace-pre-wrap">{message.content}</p>
                      
                      {/* Response Metadata */}
                      {message.response && (
                        <>
                          <Divider />
                          <div className="space-y-4">
                            {/* Confidence and Processing Time */}
                            <div className="flex items-center gap-4 text-sm text-default-500">
                              <div className="flex items-center gap-1">
                                <IconCheck className="w-4 h-4" />
                                <span>Confidence: {formatConfidenceScore(message.response.confidence)}</span>
                              </div>
                              <div className="flex items-center gap-1">
                                <IconClock className="w-4 h-4" />
                                <span>{message.response.processingTime}ms</span>
                              </div>
                              {message.response.noHallucination && (
                                <Chip size="sm" color="success" variant="flat">
                                  Source-based
                                </Chip>
                              )}
                            </div>

                            {/* Sources */}
                            {message.response.sources && message.response.sources.length > 0 && (
                              <div>
                                <h4 className="font-medium text-default-700 mb-2">Sources</h4>
                                <div className="space-y-2">
                                  {message.response.sources.map((source, index) => (
                                    <SourceCard key={index} source={source} index={index + 1} />
                                  ))}
                                </div>
                              </div>
                            )}
                          </div>
                        </>
                      )}
                    </div>
                  )}

                  {/* Message Actions */}
                  {!message.isLoading && !message.error && (
                    <div className="flex items-center justify-between mt-3 pt-3 border-t border-default-200">
                      <span className="text-xs text-default-400">
                        {formatDateTime(message.timestamp)}
                      </span>
                      <div className="flex items-center gap-1">
                        <Tooltip content="Copy response">
                          <Button
                            isIconOnly
                            size="sm"
                            variant="light"
                            onClick={() => copyToClipboard(message.content)}
                          >
                            <IconCopy className="w-4 h-4" />
                          </Button>
                        </Tooltip>
                      </div>
                    </div>
                  )}
                </CardBody>
              </Card>
            </div>

            {message.type === 'user' && (
              <Avatar
                icon={<IconUser className="w-5 h-5" />}
                className="bg-default-100 text-default-700 flex-shrink-0"
                size="sm"
              />
            )}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Query Input */}
      <div className="border-t border-default-200 p-4">
        <form onSubmit={handleSubmit} className="flex gap-3">
          <div className="flex-1">
            <Textarea
              ref={textareaRef}
              value={currentQuery}
              onChange={(e) => setCurrentQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask about your solar site... (Press Enter to send, Shift+Enter for new line)"
              minRows={1}
              maxRows={4}
              classNames={{
                input: "resize-none",
              }}
            />
          </div>
          <Button
            type="submit"
            isIconOnly
            color="primary"
            size="lg"
            isDisabled={!currentQuery.trim() || isLoading}
            className="self-end"
          >
            <IconSend className="w-5 h-5" />
          </Button>
        </form>
      </div>
    </div>
  );
}

function SourceCard({ source, index }: { source: SourceAttribution; index: number }) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <Card className="border border-default-200">
      <CardBody className="p-3">
        <div className="flex items-start gap-3">
          <div className="bg-primary/10 text-primary rounded-full w-6 h-6 flex items-center justify-center text-xs font-medium flex-shrink-0">
            {index}
          </div>
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h5 className="font-medium text-sm truncate">{source.documentTitle}</h5>
              {source.confidence && (
                <Chip size="sm" variant="flat" color="primary">
                  {formatConfidenceScore(source.confidence)}
                </Chip>
              )}
            </div>
            
            <p className="text-xs text-default-500 mb-2">{source.citation}</p>
            
            <div
              className={cn(
                'text-sm text-default-600',
                !isExpanded && 'line-clamp-2'
              )}
            >
              {source.relevantExcerpt}
            </div>
            
            {source.relevantExcerpt.length > 100 && (
              <Button
                size="sm"
                variant="light"
                color="primary"
                onClick={() => setIsExpanded(!isExpanded)}
                className="mt-1 p-0 h-auto min-w-0 text-xs"
              >
                {isExpanded ? 'Show less' : 'Show more'}
              </Button>
            )}
          </div>
          
          <Tooltip content="Open document">
            <Button
              isIconOnly
              size="sm"
              variant="light"
              onClick={() => {/* Navigate to document */}}
            >
              <IconExternalLink className="w-4 h-4" />
            </Button>
          </Tooltip>
        </div>
      </CardBody>
    </Card>
  );
}