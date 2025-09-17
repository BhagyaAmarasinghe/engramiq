'use client';

import { useState } from 'react';
import {
  Button,
  Divider,
  Chip,
  Card,
  CardBody,
  Avatar
} from '@heroui/react';
import {
  IconDashboard,
  IconFileText,
  IconMessageCircle,
  IconTool,
  IconCalendarTime,
  IconSearch,
  IconSettings,
  IconChevronLeft,
  IconChevronRight,
  IconSolarPanel,
  IconBolt,
  IconUser
} from '@tabler/icons-react';
import { EngramIQLogo } from '@/components/ui/EngramIQLogo';
import { cn } from '@/lib/utils';

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  className?: string;
}

interface NavigationItem {
  id: string;
  label: string;
  icon: React.ReactNode;
  badge?: string | number;
  description?: string;
}

const navigationItems: NavigationItem[] = [
  {
    id: 'overview',
    label: 'Overview',
    icon: <IconDashboard className="w-5 h-5" />,
    description: 'Site overview and key metrics'
  },
  {
    id: 'query',
    label: 'Ask EngramIQ',
    icon: <IconMessageCircle className="w-5 h-5" />,
    description: 'Natural language queries'
  },
  {
    id: 'documents',
    label: 'Documents',
    icon: <IconFileText className="w-5 h-5" />,
    badge: 'New',
    description: 'Upload and manage site documents'
  },
  {
    id: 'components',
    label: 'Components',
    icon: <IconSolarPanel className="w-5 h-5" />,
    description: 'Solar equipment and status'
  },
  {
    id: 'timeline',
    label: 'Timeline',
    icon: <IconCalendarTime className="w-5 h-5" />,
    description: 'Maintenance and event history'
  },
  {
    id: 'actions',
    label: 'Actions',
    icon: <IconTool className="w-5 h-5" />,
    description: 'Maintenance and repair activities'
  },
  {
    id: 'search',
    label: 'Search',
    icon: <IconSearch className="w-5 h-5" />,
    description: 'Advanced search capabilities'
  }
];

const bottomItems: NavigationItem[] = [
  {
    id: 'settings',
    label: 'Settings',
    icon: <IconSettings className="w-5 h-5" />,
    description: 'App settings and preferences'
  }
];

export function Sidebar({ activeTab, onTabChange, className }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className={cn(
      'flex flex-col h-full bg-primary-dark-blue border-r border-default-200 transition-all duration-300',
      collapsed ? 'w-16' : 'w-64',
      className
    )}>
      {/* Header */}
      <div className="p-4 border-b border-default-200">
        <div className="flex items-center justify-between">
          <EngramIQLogo 
            size={collapsed ? 'sm' : 'md'} 
            showText={!collapsed}
            variant="white"
          />
          <Button
            isIconOnly
            size="sm"
            variant="light"
            onClick={() => setCollapsed(!collapsed)}
            className="text-white hover:bg-white/10"
          >
            {collapsed ? (
              <IconChevronRight className="w-4 h-4" />
            ) : (
              <IconChevronLeft className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>

      {/* Site Info */}
      {!collapsed && (
        <div className="p-4 border-b border-default-200">
          <Card className="glass-effect">
            <CardBody className="p-3">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-primary/10 rounded-lg">
                  <IconBolt className="w-4 h-4 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="text-sm font-medium text-white truncate">Site S2367</h3>
                  <p className="text-xs text-default-400">2,850 kW • 36 Inverters</p>
                </div>
                <Chip size="sm" color="success" variant="flat">Online</Chip>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 p-2 space-y-1">
        {navigationItems.map((item) => (
          <NavigationButton
            key={item.id}
            item={item}
            isActive={activeTab === item.id}
            collapsed={collapsed}
            onClick={() => onTabChange(item.id)}
          />
        ))}
      </nav>

      {/* Bottom Section */}
      <div className="border-t border-default-200 p-2 space-y-1">
        {bottomItems.map((item) => (
          <NavigationButton
            key={item.id}
            item={item}
            isActive={activeTab === item.id}
            collapsed={collapsed}
            onClick={() => onTabChange(item.id)}
          />
        ))}
        
        <Divider />
        
        {/* User Profile */}
        <div className={cn(
          'p-2 rounded-lg hover:bg-white/5 transition-colors cursor-pointer',
          collapsed ? 'justify-center' : ''
        )}>
          <div className="flex items-center gap-3">
            <Avatar
              name="User"
              size="sm"
              className="bg-primary text-white"
              icon={<IconUser className="w-4 h-4" />}
            />
            {!collapsed && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white">Site Manager</p>
                <p className="text-xs text-default-400 truncate">user@engramiq.com</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function NavigationButton({ 
  item, 
  isActive, 
  collapsed, 
  onClick 
}: { 
  item: NavigationItem;
  isActive: boolean;
  collapsed: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      className={cn(
        'w-full justify-start gap-3 h-auto p-3 font-normal transition-all duration-200',
        collapsed ? 'min-w-0 px-2' : '',
        isActive 
          ? 'bg-primary/20 text-primary border-primary/30' 
          : 'text-white/80 hover:text-white hover:bg-white/5'
      )}
      variant={isActive ? 'flat' : 'light'}
      onClick={onClick}
      startContent={collapsed ? undefined : item.icon}
    >
      {collapsed ? (
        <div className="flex justify-center">
          {item.icon}
        </div>
      ) : (
        <>
          <div className="flex-1 text-left">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">{item.label}</span>
              {item.badge && (
                <Chip size="sm" color="primary" variant="flat">
                  {item.badge}
                </Chip>
              )}
            </div>
            {item.description && (
              <p className="text-xs text-default-400 mt-0.5">{item.description}</p>
            )}
          </div>
        </>
      )}
    </Button>
  );
}