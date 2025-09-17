'use client';

import { useState } from 'react';
import {
  Button,
  Divider,
  Card,
  CardBody,
  Avatar,
  Listbox,
  ListboxItem,
  ListboxSection,
  cn
} from '@heroui/react';
import {Chip} from '@heroui/chip';
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
                <Chip
                  size="sm"
                  color="success"
                  variant="flat"
                  startContent={
                    <div className="w-2 h-2 bg-success rounded-full animate-pulse" />
                  }
                >
                  Online
                </Chip>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      {/* Navigation with HeroUI Listbox */}
      {!collapsed && (
        <div className="flex-1 p-4">
          <Listbox
            aria-label="Navigation Menu"
            className="p-0 gap-0 bg-transparent"
            selectedKeys={new Set([activeTab])}
            selectionMode="single"
            onSelectionChange={(keys) => {
              const selectedKey = Array.from(keys)[0];
              if (selectedKey) {
                onTabChange(selectedKey as string);
              }
            }}
            itemClasses={{
              base: "px-3 first:rounded-t-medium last:rounded-b-medium rounded-medium gap-3 h-12 data-[hover=true]:bg-white/5 data-[selected=true]:bg-primary/10 data-[selected=true]:text-primary mb-1",
            }}
          >
            <ListboxSection title="Main Navigation" showDivider>
              <ListboxItem
                key="overview"
                description="Site overview and key metrics"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'overview' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconDashboard className="w-4 h-4" />
                  </div>
                }
              >
                Overview
              </ListboxItem>
              <ListboxItem
                key="query"
                description="Natural language queries"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'query' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconMessageCircle className="w-4 h-4" />
                  </div>
                }
              >
                Ask EngramIQ
              </ListboxItem>
              <ListboxItem
                key="documents"
                description="Upload and manage site documents"
                endContent={
                  <Chip
                    size="sm"
                    color="primary"
                    variant="flat"
                    startContent={
                      <div className="w-1.5 h-1.5 bg-primary rounded-full" />
                    }
                  >
                    New
                  </Chip>
                }
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'documents' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconFileText className="w-4 h-4" />
                  </div>
                }
              >
                Documents
              </ListboxItem>
              <ListboxItem
                key="components"
                description="Solar equipment and status"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'components' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconSolarPanel className="w-4 h-4" />
                  </div>
                }
              >
                Components
              </ListboxItem>
              <ListboxItem
                key="timeline"
                description="Maintenance and event history"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'timeline' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconCalendarTime className="w-4 h-4" />
                  </div>
                }
              >
                Timeline
              </ListboxItem>
              <ListboxItem
                key="actions"
                description="Maintenance and repair activities"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'actions' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconTool className="w-4 h-4" />
                  </div>
                }
              >
                Actions
              </ListboxItem>
              <ListboxItem
                key="search"
                description="Advanced search capabilities"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'search' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconSearch className="w-4 h-4" />
                  </div>
                }
              >
                Search
              </ListboxItem>
            </ListboxSection>

            <ListboxSection title="System">
              <ListboxItem
                key="settings"
                description="App settings and preferences"
                startContent={
                  <div className={cn("flex items-center justify-center w-8 h-8 rounded-lg",
                    activeTab === 'settings' ? 'bg-primary/20 text-primary' : 'bg-white/10 text-white')}>
                    <IconSettings className="w-4 h-4" />
                  </div>
                }
              >
                Settings
              </ListboxItem>
            </ListboxSection>
          </Listbox>
        </div>
      )}

      {/* Collapsed Navigation */}
      {collapsed && (
        <nav className="flex-1 p-2 space-y-2">
          {[...navigationItems, ...bottomItems].map((item) => (
            <Button
              key={item.id}
              isIconOnly
              variant="light"
              className={cn(
                'w-full h-12 text-white/80 hover:text-white hover:bg-white/5',
                activeTab === item.id && 'bg-primary/20 text-primary'
              )}
              onClick={() => onTabChange(item.id)}
            >
              {item.icon}
            </Button>
          ))}
        </nav>
      )}

      {/* User Profile */}
      <div className="border-t border-default-200 p-4">
        <div className={cn(
          'flex items-center gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors cursor-pointer',
          collapsed && 'justify-center'
        )}>
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
  );
}

