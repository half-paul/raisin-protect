'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { GrcRole, getRoleLabel } from '@/lib/auth';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import {
  LayoutDashboard,
  Shield,
  FileCheck,
  AlertTriangle,
  Monitor,
  Settings,
  Users,
  Building2,
  Code2,
  ClipboardCheck,
  Truck,
  LogOut,
  ChevronLeft,
  X,
  Grid3X3,
  BarChart3,
  Activity,
  Bell,
  PlayCircle,
  Settings2,
  FileText,
  BookOpen,
  Target,
  CheckSquare,
} from 'lucide-react';

interface NavItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  roles?: GrcRole[]; // if undefined, visible to all roles
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const navigation: NavSection[] = [
  {
    title: 'Overview',
    items: [
      { label: 'Dashboard', href: '/', icon: LayoutDashboard },
    ],
  },
  {
    title: 'Risk & Posture',
    items: [
      {
        label: 'Risk Dashboard',
        href: '/risk',
        icon: AlertTriangle,
        roles: ['ciso', 'compliance_manager', 'security_engineer'],
      },
      {
        label: 'Posture Overview',
        href: '/posture',
        icon: Shield,
        roles: ['ciso', 'compliance_manager', 'security_engineer'],
      },
    ],
  },
  {
    title: 'Compliance',
    items: [
      {
        label: 'Frameworks',
        href: '/frameworks',
        icon: FileCheck,
        roles: ['ciso', 'compliance_manager', 'auditor'],
      },
      {
        label: 'Controls',
        href: '/controls',
        icon: Shield,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
      {
        label: 'Mapping Matrix',
        href: '/mapping-matrix',
        icon: Grid3X3,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
      {
        label: 'Coverage',
        href: '/coverage',
        icon: BarChart3,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
      {
        label: 'Evidence',
        href: '/evidence',
        icon: ClipboardCheck,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
      {
        label: 'Staleness Alerts',
        href: '/staleness',
        icon: AlertTriangle,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
    ],
  },
  {
    title: 'Policy Management',
    items: [
      {
        label: 'Policies',
        href: '/policies',
        icon: FileText,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
      {
        label: 'Templates',
        href: '/policy-templates',
        icon: BookOpen,
        roles: ['ciso', 'compliance_manager', 'security_engineer'],
      },
      {
        label: 'Approvals',
        href: '/policy-approvals',
        icon: CheckSquare,
      },
      {
        label: 'Policy Gaps',
        href: '/policy-gap',
        icon: Target,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'auditor'],
      },
    ],
  },
  {
    title: 'Monitoring',
    items: [
      {
        label: 'Monitoring',
        href: '/monitoring',
        icon: Activity,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'devops_engineer'],
      },
      {
        label: 'Alert Queue',
        href: '/alerts',
        icon: Bell,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'devops_engineer', 'it_admin'],
      },
      {
        label: 'Test Runs',
        href: '/test-runs',
        icon: PlayCircle,
        roles: ['ciso', 'compliance_manager', 'security_engineer', 'devops_engineer'],
      },
      {
        label: 'Alert Rules',
        href: '/alert-rules',
        icon: Settings2,
        roles: ['ciso', 'compliance_manager', 'security_engineer'],
      },
    ],
  },
  {
    title: 'Audit',
    items: [
      {
        label: 'Audit Hub',
        href: '/audit',
        icon: ClipboardCheck,
        roles: ['ciso', 'compliance_manager', 'auditor'],
      },
    ],
  },
  {
    title: 'Vendor Management',
    items: [
      {
        label: 'Vendors',
        href: '/vendors',
        icon: Truck,
        roles: ['ciso', 'compliance_manager', 'vendor_manager'],
      },
    ],
  },
  {
    title: 'Integration',
    items: [
      {
        label: 'Integrations',
        href: '/integrations',
        icon: Code2,
        roles: ['ciso', 'compliance_manager', 'it_admin', 'devops_engineer'],
      },
      {
        label: 'API',
        href: '/api-docs',
        icon: Code2,
        roles: ['devops_engineer'],
      },
    ],
  },
  {
    title: 'Administration',
    items: [
      {
        label: 'Users',
        href: '/users',
        icon: Users,
        roles: ['ciso', 'compliance_manager', 'it_admin'],
      },
      {
        label: 'Organization',
        href: '/settings',
        icon: Building2,
        roles: ['ciso', 'compliance_manager'],
      },
    ],
  },
];

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  const filteredNavigation = navigation
    .map((section) => ({
      ...section,
      items: section.items.filter(
        (item) => !item.roles || (user && item.roles.includes(user.role))
      ),
    }))
    .filter((section) => section.items.length > 0);

  const initials = user
    ? `${user.first_name?.[0] || ''}${user.last_name?.[0] || ''}`.toUpperCase() || user.email[0].toUpperCase()
    : '?';

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={onClose}
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 border-r bg-card flex flex-col transition-transform duration-200 md:relative md:translate-x-0',
          isOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            <Shield className="h-6 w-6 text-primary" />
            <span className="font-bold text-lg">Raisin Protect</span>
          </div>
          <Button variant="ghost" size="icon" className="md:hidden" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-4 px-3">
          {filteredNavigation.map((section) => (
            <div key={section.title} className="mb-4">
              <p className="px-3 mb-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {section.title}
              </p>
              {section.items.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={onClose}
                    className={cn(
                      'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-primary/10 text-primary'
                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                    )}
                  >
                    <item.icon className="h-4 w-4" />
                    {item.label}
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>

        {/* User section */}
        <div className="border-t p-4">
          <div className="flex items-center gap-3 mb-3">
            <Avatar className="h-8 w-8">
              <AvatarFallback className="text-xs">{initials}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">
                {user?.first_name} {user?.last_name}
              </p>
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                {user?.role ? getRoleLabel(user.role) : 'Unknown'}
              </Badge>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-muted-foreground"
            onClick={() => logout()}
          >
            <LogOut className="h-4 w-4 mr-2" />
            Sign out
          </Button>
        </div>
      </aside>
    </>
  );
}
