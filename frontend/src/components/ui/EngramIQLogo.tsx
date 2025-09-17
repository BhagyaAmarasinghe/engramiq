'use client';

import { cn } from '@/lib/utils';

interface EngramIQLogoProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  variant?: 'default' | 'gradient' | 'white';
  showText?: boolean;
  className?: string;
}

export function EngramIQLogo({ 
  size = 'md', 
  variant = 'default', 
  showText = true,
  className 
}: EngramIQLogoProps) {
  const sizeClasses = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
    xl: 'w-16 h-16'
  };

  const textSizeClasses = {
    sm: 'text-lg',
    md: 'text-xl',
    lg: 'text-2xl',
    xl: 'text-3xl'
  };

  const getIconColor = () => {
    switch (variant) {
      case 'gradient':
        return 'bg-brand-gradient';
      case 'white':
        return 'text-white';
      default:
        return 'text-primary-green';
    }
  };

  const getTextColor = () => {
    switch (variant) {
      case 'gradient':
        return 'gradient-text';
      case 'white':
        return 'text-white';
      default:
        return 'text-white';
    }
  };

  return (
    <div className={cn('flex items-center gap-3', className)}>
      {/* Logo Icon */}
      <div className={cn(sizeClasses[size], 'relative')}>
        <svg 
          viewBox="0 0 40 40" 
          fill="none" 
          xmlns="http://www.w3.org/2000/svg"
          className={cn('w-full h-full', variant === 'gradient' ? '' : getIconColor())}
        >
          {variant === 'gradient' ? (
            <>
              <defs>
                <linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stopColor="#17c480" />
                  <stop offset="100%" stopColor="#0d1830" />
                </linearGradient>
              </defs>
              <g fill="url(#logo-gradient)">
                {/* Outer ring segments */}
                <path d="M20 2 L32 8 L32 16 L20 10 Z" />
                <path d="M32 16 L32 24 L20 30 L20 22 Z" />
                <path d="M20 30 L8 24 L8 16 L20 22 Z" />
                <path d="M8 16 L8 8 L20 2 L20 10 Z" />
                
                {/* Inner circle */}
                <circle cx="20" cy="20" r="6" />
                
                {/* Connection lines */}
                <rect x="14" y="19" width="12" height="2" />
                <rect x="19" y="14" width="2" height="12" />
              </g>
            </>
          ) : (
            <g fill="currentColor">
              {/* Outer ring segments */}
              <path d="M20 2 L32 8 L32 16 L20 10 Z" />
              <path d="M32 16 L32 24 L20 30 L20 22 Z" />
              <path d="M20 30 L8 24 L8 16 L20 22 Z" />
              <path d="M8 16 L8 8 L20 2 L20 10 Z" />
              
              {/* Inner circle */}
              <circle cx="20" cy="20" r="6" />
              
              {/* Connection lines */}
              <rect x="14" y="19" width="12" height="2" />
              <rect x="19" y="14" width="2" height="12" />
            </g>
          )}
        </svg>
      </div>

      {/* Logo Text */}
      {showText && (
        <span className={cn(
          'font-bold tracking-wide font-figtree',
          textSizeClasses[size],
          getTextColor()
        )}>
          ENGRAMIQ
        </span>
      )}
    </div>
  );
}