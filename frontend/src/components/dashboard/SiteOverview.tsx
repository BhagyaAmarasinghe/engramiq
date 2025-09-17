'use client';

import { useState, useEffect } from 'react';
import {
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Progress,
  Button,
  Skeleton,
  Divider,
  Avatar,
  Spacer
} from '@heroui/react';
import {Chip} from '@heroui/chip';
import {
  IconSolarPanel,
  IconActivity,
  IconFileText,
  IconAlertTriangle,
  IconCheck,
  IconClock,
  IconTool,
  IconTrendingUp
} from '@tabler/icons-react';
import { cn, formatDate, getStatusColor } from '@/lib/utils';
import { Site, Component, Document, ExtractedAction } from '@/types';

interface SiteOverviewProps {
  site: Site;
  components: Component[];
  documents: Document[];
  actions: ExtractedAction[];
  isLoading?: boolean;
  className?: string;
}

interface StatsCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'danger' | 'default';
  trend?: {
    value: number;
    isPositive: boolean;
  };
}

function StatsCard({ title, value, subtitle, icon, color = 'primary', trend }: StatsCardProps) {
  return (
    <Card className="glass-effect min-w-[280px] flex-1">
      <CardBody className="p-6">
        {/* Header with icon and title */}
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className={cn(
              'flex items-center justify-center w-12 h-12 rounded-xl',
              color === 'primary' && 'bg-primary/15 text-primary',
              color === 'success' && 'bg-success/15 text-success',
              color === 'warning' && 'bg-warning/15 text-warning',
              color === 'danger' && 'bg-danger/15 text-danger',
              color === 'default' && 'bg-default/15 text-default-500'
            )}>
              {icon}
            </div>
            <div>
              <h3 className="text-sm font-medium text-default-500 uppercase tracking-wide">{title}</h3>
            </div>
          </div>
        </div>

        {/* Main value */}
        <div className="mb-4">
          <div className="text-3xl font-bold text-white mb-1">{value}</div>
          {subtitle && (
            <p className="text-sm text-default-400">{subtitle}</p>
          )}
        </div>

        {/* Trend indicator */}
        {trend && (
          <div className="flex items-center gap-2 pt-3 border-t border-default-200/30">
            <div className="flex items-center gap-1">
              <IconTrendingUp className={cn(
                'w-4 h-4',
                trend.isPositive ? 'text-success rotate-0' : 'text-danger rotate-180'
              )} />
              <span className={cn(
                'text-sm font-semibold',
                trend.isPositive ? 'text-success' : 'text-danger'
              )}>
                {Math.abs(trend.value)}%
              </span>
            </div>
            <span className="text-sm text-default-400">vs last month</span>
          </div>
        )}
      </CardBody>
    </Card>
  );
}

export function SiteOverview({ 
  site, 
  components, 
  documents, 
  actions,
  isLoading,
  className 
}: SiteOverviewProps) {
  const [componentStats, setComponentStats] = useState({
    total: 0,
    operational: 0,
    fault: 0,
    maintenance: 0,
    offline: 0
  });

  const [documentStats, setDocumentStats] = useState({
    total: 0,
    processed: 0,
    pending: 0,
    failed: 0
  });

  const [actionStats, setActionStats] = useState({
    total: 0,
    thisMonth: 0,
    completed: 0,
    pending: 0
  });

  useEffect(() => {
    // Calculate component statistics - with defensive check
    const componentCounts = (components || []).reduce((acc, component) => {
      acc.total += 1;
      acc[component.currentStatus] = (acc[component.currentStatus] || 0) + 1;
      return acc;
    }, { 
      total: 0, 
      operational: 0, 
      fault: 0, 
      maintenance: 0, 
      offline: 0 
    });

    setComponentStats(componentCounts);

    // Calculate document statistics - with defensive check
    const documentCounts = (documents || []).reduce((acc, document) => {
      acc.total += 1;
      acc[document.processingStatus] = (acc[document.processingStatus] || 0) + 1;
      return acc;
    }, { 
      total: 0, 
      completed: 0, 
      pending: 0, 
      processing: 0,
      failed: 0 
    });

    setDocumentStats({
      total: documentCounts.total,
      processed: documentCounts.completed,
      pending: documentCounts.pending + documentCounts.processing,
      failed: documentCounts.failed
    });

    // Calculate action statistics - with defensive check
    const currentMonth = new Date().getMonth();
    const currentYear = new Date().getFullYear();
    
    const actionsArray = actions || [];
    const thisMonthActions = actionsArray.filter(action => {
      const actionDate = new Date(action.createdAt);
      return actionDate.getMonth() === currentMonth && actionDate.getFullYear() === currentYear;
    });

    setActionStats({
      total: actionsArray.length,
      thisMonth: thisMonthActions.length,
      completed: actionsArray.filter(a => a.status === 'completed').length,
      pending: actionsArray.filter(a => ['planned', 'in_progress'].includes(a.status)).length
    });
  }, [components, documents, actions]);

  if (isLoading) {
    return (
      <div className={cn('space-y-6', className)}>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[...Array(4)].map((_, i) => (
            <Card key={i} className="glass-effect">
              <CardBody className="p-6">
                <Skeleton className="rounded-lg mb-4">
                  <div className="h-16 rounded-lg bg-default-200"></div>
                </Skeleton>
                <Skeleton className="w-3/5 rounded-lg mb-2">
                  <div className="h-4 rounded-lg bg-default-200"></div>
                </Skeleton>
                <Skeleton className="w-4/5 rounded-lg">
                  <div className="h-3 rounded-lg bg-default-200"></div>
                </Skeleton>
              </CardBody>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  const operationalPercentage = componentStats.total > 0 
    ? (componentStats.operational / componentStats.total * 100).toFixed(1)
    : '0';

  const processingPercentage = documentStats.total > 0
    ? (documentStats.processed / documentStats.total * 100).toFixed(1)
    : '0';

  return (
    <div className={cn('space-y-6', className)}>
      {/* Site Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">{site.name}</h1>
          <div className="flex items-center gap-4 text-default-400">
            <span>{site.site_code}</span>
            <span>•</span>
            <span>{(site.total_capacity_kw || 0).toLocaleString()} kW</span>
            <span>•</span>
            <span>{site.number_of_inverters || 0} inverters</span>
            {site.installation_date && (
              <>
                <span>•</span>
                <span>Installed {formatDate(site.installation_date)}</span>
              </>
            )}
          </div>
        </div>
        
        <Chip
          color="success"
          variant="flat"
          size="lg"
          startContent={
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 bg-success rounded-full animate-pulse" />
              <IconCheck className="w-4 h-4" />
            </div>
          }
        >
          Online
        </Chip>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        <StatsCard
          title="System Health"
          value={`${operationalPercentage}%`}
          subtitle={`${componentStats.operational} of ${componentStats.total} operational`}
          icon={<IconSolarPanel className="w-5 h-5" />}
          color={componentStats.fault > 0 ? 'warning' : 'success'}
          trend={{ value: 2.5, isPositive: true }}
        />

        <StatsCard
          title="Documents Processed"
          value={documentStats.processed}
          subtitle={`${processingPercentage}% completion rate`}
          icon={<IconFileText className="w-5 h-5" />}
          color="primary"
          trend={{ value: 12, isPositive: true }}
        />

        <StatsCard
          title="Actions This Month"
          value={actionStats.thisMonth}
          subtitle={`${actionStats.completed} completed`}
          icon={<IconTool className="w-5 h-5" />}
          color="success"
          trend={{ value: 8, isPositive: false }}
        />

        <StatsCard
          title="Active Alerts"
          value={componentStats.fault + componentStats.offline}
          subtitle={componentStats.fault ? `${componentStats.fault} faults` : 'All systems normal'}
          icon={<IconAlertTriangle className="w-5 h-5" />}
          color={componentStats.fault > 0 ? 'danger' : 'success'}
        />
      </div>

      {/* System Status Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Component Status */}
        <Card className="glass-effect">
          <CardHeader>
            <h3 className="text-lg font-semibold">Component Status</h3>
          </CardHeader>
          <CardBody className="pt-0">
            <div className="space-y-4">
              <Progress
                label="Operational Components"
                value={Number(operationalPercentage)}
                color="success"
                showValueLabel
                classNames={{
                  label: "text-white",
                  value: "text-success"
                }}
              />
              
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-default-500">Operational</span>
                    <Chip size="sm" color="success" variant="flat">
                      {componentStats.operational}
                    </Chip>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-default-500">Maintenance</span>
                    <Chip size="sm" color="warning" variant="flat">
                      {componentStats.maintenance}
                    </Chip>
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-default-500">Fault</span>
                    <Chip size="sm" color="danger" variant="flat">
                      {componentStats.fault}
                    </Chip>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-default-500">Offline</span>
                    <Chip size="sm" color="default" variant="flat">
                      {componentStats.offline}
                    </Chip>
                  </div>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Recent Activity Summary */}
        <Card className="glass-effect">
          <CardHeader className="flex justify-between items-center">
            <h3 className="text-lg font-semibold">Recent Activity</h3>
            <Button size="sm" variant="flat" color="primary">
              View All
            </Button>
          </CardHeader>
          <CardBody className="pt-0">
            <div className="space-y-3">
              <div className="flex items-center gap-3 p-3 rounded-lg bg-default-100/10">
                <IconActivity className="w-5 h-5 text-primary" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">
                    {actionStats.thisMonth} maintenance actions completed
                  </p>
                  <p className="text-xs text-default-500">This month</p>
                </div>
              </div>
              
              <div className="flex items-center gap-3 p-3 rounded-lg bg-default-100/10">
                <IconFileText className="w-5 h-5 text-success" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">
                    {documentStats.processed} documents processed
                  </p>
                  <p className="text-xs text-default-500">
                    {documentStats.pending} pending processing
                  </p>
                </div>
              </div>
              
              {componentStats.fault > 0 && (
                <div className="flex items-center gap-3 p-3 rounded-lg bg-danger-50/10">
                  <IconAlertTriangle className="w-5 h-5 text-danger" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">
                      {componentStats.fault} component{componentStats.fault !== 1 ? 's' : ''} need{componentStats.fault === 1 ? 's' : ''} attention
                    </p>
                    <p className="text-xs text-danger">Requires immediate action</p>
                  </div>
                </div>
              )}
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}