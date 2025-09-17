'use client';

import { useState, useMemo } from 'react';
import {
  Card,
  CardHeader,
  CardBody,
  Button,
  Input,
  Select,
  SelectItem,
  DateRangePicker,
  Chip,
  Avatar,
  Skeleton,
  Pagination
} from '@heroui/react';
import {
  IconCalendar,
  IconFilter,
  IconSearch,
  IconClock,
  IconUser,
  IconTool,
  IconAlertTriangle,
  IconCheck,
  IconFileText
} from '@tabler/icons-react';
import { cn, formatDateTime, formatDate, getStatusColor } from '@/lib/utils';
import { TimelineEvent, EventType, EventPriority } from '@/types';

interface TimelineProps {
  events: TimelineEvent[];
  isLoading?: boolean;
  className?: string;
}

interface TimelineFilters {
  search: string;
  eventType: EventType | 'all';
  priority: EventPriority | 'all';
  dateRange: { start: Date | null; end: Date | null };
}

const eventTypeOptions = [
  { key: 'all', label: 'All Events' },
  { key: 'maintenance_scheduled', label: 'Maintenance Scheduled' },
  { key: 'maintenance_completed', label: 'Maintenance Completed' },
  { key: 'fault_occurred', label: 'Fault Occurred' },
  { key: 'fault_cleared', label: 'Fault Cleared' },
  { key: 'replacement_scheduled', label: 'Replacement Scheduled' },
  { key: 'replacement_completed', label: 'Replacement Completed' },
  { key: 'inspection_scheduled', label: 'Inspection Scheduled' },
  { key: 'inspection_completed', label: 'Inspection Completed' },
  { key: 'warranty_claim', label: 'Warranty Claim' },
  { key: 'performance_alert', label: 'Performance Alert' },
];

const priorityOptions = [
  { key: 'all', label: 'All Priorities' },
  { key: 'critical', label: 'Critical' },
  { key: 'high', label: 'High' },
  { key: 'medium', label: 'Medium' },
  { key: 'low', label: 'Low' },
];

function getEventIcon(eventType: EventType) {
  switch (eventType) {
    case 'maintenance_scheduled':
    case 'maintenance_completed':
      return <IconTool className="w-4 h-4" />;
    case 'fault_occurred':
      return <IconAlertTriangle className="w-4 h-4" />;
    case 'fault_cleared':
      return <IconCheck className="w-4 h-4" />;
    case 'inspection_scheduled':
    case 'inspection_completed':
      return <IconSearch className="w-4 h-4" />;
    case 'replacement_scheduled':
    case 'replacement_completed':
      return <IconTool className="w-4 h-4" />;
    case 'warranty_claim':
      return <IconFileText className="w-4 h-4" />;
    default:
      return <IconClock className="w-4 h-4" />;
  }
}

function getEventTypeColor(eventType: EventType): 'primary' | 'secondary' | 'default' | 'success' | 'warning' | 'danger' {
  switch (eventType) {
    case 'maintenance_completed':
    case 'fault_cleared':
    case 'inspection_completed':
    case 'replacement_completed':
      return 'success';
    case 'fault_occurred':
    case 'performance_alert':
      return 'danger';
    case 'maintenance_scheduled':
    case 'inspection_scheduled':
    case 'replacement_scheduled':
      return 'warning';
    case 'warranty_claim':
      return 'secondary';
    default:
      return 'default';
  }
}

function formatEventType(eventType: EventType): string {
  return eventType.split('_').map(word => 
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ');
}

function TimelineEventCard({ event }: { event: TimelineEvent }) {
  const eventIcon = getEventIcon(event.eventType);
  const eventColor = getEventTypeColor(event.eventType);

  return (
    <div className="flex gap-4">
      {/* Timeline indicator */}
      <div className="flex flex-col items-center">
        <div className={cn(
          'p-2 rounded-full border-2',
          eventColor === 'success' && 'bg-success/10 border-success text-success',
          eventColor === 'danger' && 'bg-danger/10 border-danger text-danger',
          eventColor === 'warning' && 'bg-warning/10 border-warning text-warning',
          eventColor === 'secondary' && 'bg-secondary/10 border-secondary text-secondary',
          eventColor === 'default' && 'bg-default/10 border-default-300 text-default-500'
        )}>
          {eventIcon}
        </div>
        <div className="w-px bg-default-200 flex-1 mt-2" />
      </div>

      {/* Event content */}
      <Card className="flex-1 glass-effect mb-4">
        <CardBody className="p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="font-semibold text-white truncate">{event.title}</h3>
                <Chip 
                  size="sm" 
                  color={getStatusColor(event.priority)} 
                  variant="flat"
                >
                  {event.priority}
                </Chip>
              </div>
              
              <div className="flex items-center gap-3 text-sm text-default-500 mb-2">
                <div className="flex items-center gap-1">
                  <IconClock className="w-4 h-4" />
                  <span>{formatDateTime(event.startTime)}</span>
                </div>
                
                {event.durationMinutes && (
                  <span>Duration: {event.durationMinutes} min</span>
                )}
              </div>

              <Chip 
                size="sm" 
                color={eventColor} 
                variant="flat"
                className="mb-3"
              >
                {formatEventType(event.eventType)}
              </Chip>
            </div>
          </div>

          {event.description && (
            <p className="text-default-300 text-sm mb-3">{event.description}</p>
          )}

          {/* Event details */}
          <div className="space-y-2">
            {event.assignedTechnicians.length > 0 && (
              <div className="flex items-center gap-2">
                <IconUser className="w-4 h-4 text-default-500" />
                <div className="flex items-center gap-1">
                  {event.assignedTechnicians.slice(0, 3).map((technician, index) => (
                    <Avatar
                      key={index}
                      name={technician}
                      size="sm"
                      className="w-6 h-6 text-xs"
                    />
                  ))}
                  {event.assignedTechnicians.length > 3 && (
                    <span className="text-xs text-default-500 ml-1">
                      +{event.assignedTechnicians.length - 3} more
                    </span>
                  )}
                </div>
              </div>
            )}

            {event.affectedComponentIds.length > 0 && (
              <div className="flex items-center gap-2">
                <IconTool className="w-4 h-4 text-default-500" />
                <span className="text-sm text-default-400">
                  {event.affectedComponentIds.length} component{event.affectedComponentIds.length !== 1 ? 's' : ''} affected
                </span>
              </div>
            )}
          </div>
        </CardBody>
      </Card>
    </div>
  );
}

export function Timeline({ events, isLoading, className }: TimelineProps) {
  const [filters, setFilters] = useState<TimelineFilters>({
    search: '',
    eventType: 'all',
    priority: 'all',
    dateRange: { start: null, end: null }
  });
  const [currentPage, setCurrentPage] = useState(1);
  const eventsPerPage = 10;

  const filteredEvents = useMemo(() => {
    return events.filter(event => {
      // Search filter
      if (filters.search && !event.title.toLowerCase().includes(filters.search.toLowerCase()) &&
          !event.description?.toLowerCase().includes(filters.search.toLowerCase())) {
        return false;
      }

      // Event type filter
      if (filters.eventType !== 'all' && event.eventType !== filters.eventType) {
        return false;
      }

      // Priority filter
      if (filters.priority !== 'all' && event.priority !== filters.priority) {
        return false;
      }

      // Date range filter
      if (filters.dateRange.start || filters.dateRange.end) {
        const eventDate = new Date(event.startTime);
        if (filters.dateRange.start && eventDate < filters.dateRange.start) {
          return false;
        }
        if (filters.dateRange.end && eventDate > filters.dateRange.end) {
          return false;
        }
      }

      return true;
    }).sort((a, b) => new Date(b.startTime).getTime() - new Date(a.startTime).getTime());
  }, [events, filters]);

  const paginatedEvents = useMemo(() => {
    const start = (currentPage - 1) * eventsPerPage;
    return filteredEvents.slice(start, start + eventsPerPage);
  }, [filteredEvents, currentPage]);

  const totalPages = Math.ceil(filteredEvents.length / eventsPerPage);

  if (isLoading) {
    return (
      <div className={cn('space-y-4', className)}>
        <Card className="glass-effect">
          <CardHeader>
            <Skeleton className="w-3/5 rounded-lg">
              <div className="h-6 rounded-lg bg-default-200"></div>
            </Skeleton>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex gap-4">
                  <Skeleton className="w-12 h-12 rounded-full" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="w-4/5 rounded-lg">
                      <div className="h-4 rounded-lg bg-default-200"></div>
                    </Skeleton>
                    <Skeleton className="w-3/5 rounded-lg">
                      <div className="h-3 rounded-lg bg-default-200"></div>
                    </Skeleton>
                  </div>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* Filters */}
      <Card className="glass-effect">
        <CardHeader>
          <div className="flex items-center justify-between w-full">
            <h2 className="text-xl font-semibold">Timeline</h2>
            <div className="flex items-center gap-2">
              <IconFilter className="w-5 h-5 text-default-500" />
              <span className="text-sm text-default-500">
                {filteredEvents.length} of {events.length} events
              </span>
            </div>
          </div>
        </CardHeader>
        <CardBody className="pt-0">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Input
              placeholder="Search events..."
              value={filters.search}
              onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
              startContent={<IconSearch className="w-4 h-4 text-default-500" />}
              size="sm"
            />

            <Select
              placeholder="Event Type"
              selectedKeys={filters.eventType ? [filters.eventType] : []}
              onSelectionChange={(keys) => {
                const type = Array.from(keys)[0] as EventType | 'all';
                setFilters(prev => ({ ...prev, eventType: type }));
              }}
              size="sm"
            >
              {eventTypeOptions.map((option) => (
                <SelectItem key={option.key}>
                  {option.label}
                </SelectItem>
              ))}
            </Select>

            <Select
              placeholder="Priority"
              selectedKeys={filters.priority ? [filters.priority] : []}
              onSelectionChange={(keys) => {
                const priority = Array.from(keys)[0] as EventPriority | 'all';
                setFilters(prev => ({ ...prev, priority }));
              }}
              size="sm"
            >
              {priorityOptions.map((option) => (
                <SelectItem key={option.key}>
                  {option.label}
                </SelectItem>
              ))}
            </Select>

            <Button
              variant="flat"
              color="primary"
              onClick={() => setFilters({
                search: '',
                eventType: 'all',
                priority: 'all',
                dateRange: { start: null, end: null }
              })}
              size="sm"
            >
              Clear Filters
            </Button>
          </div>
        </CardBody>
      </Card>

      {/* Timeline Events */}
      <div className="relative">
        {paginatedEvents.length === 0 ? (
          <Card className="glass-effect">
            <CardBody className="text-center py-12">
              <IconCalendar className="w-12 h-12 text-default-300 mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-default-500 mb-2">No events found</h3>
              <p className="text-default-400">
                {filters.search || filters.eventType !== 'all' || filters.priority !== 'all'
                  ? 'Try adjusting your filters to see more events.'
                  : 'No timeline events available for this site.'
                }
              </p>
            </CardBody>
          </Card>
        ) : (
          <>
            <div className="space-y-0">
              {paginatedEvents.map((event, index) => (
                <TimelineEventCard 
                  key={event.id} 
                  event={event}
                />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex justify-center mt-8">
                <Pagination
                  total={totalPages}
                  page={currentPage}
                  onChange={setCurrentPage}
                  color="primary"
                  showControls
                />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}