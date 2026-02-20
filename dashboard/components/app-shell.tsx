'use client';

import { useState } from 'react';
import { usePathname } from 'next/navigation';
import { Sidebar } from '@/components/sidebar';
import { MobileHeader } from '@/components/mobile-header';

const BYPASS_SHELL_PATHS = ['/login', '/register'];

function shouldBypassShell(pathname: string | null): boolean {
  if (!pathname) return false;
  return BYPASS_SHELL_PATHS.some((p) => pathname === p || pathname.startsWith(p + '/'));
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();

  if (shouldBypassShell(pathname)) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen flex-col md:flex-row">
      <MobileHeader onMenuClick={() => setSidebarOpen(true)} />
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      <main className="flex-1 overflow-auto bg-background">
        {children}
      </main>
    </div>
  );
}
