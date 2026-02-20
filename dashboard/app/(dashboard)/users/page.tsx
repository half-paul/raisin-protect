'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { authFetch, GrcRole, getRoleLabel, User } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  Users,
  Plus,
  Search,
  UserCog,
  UserMinus,
  UserCheck,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

const ALL_ROLES: GrcRole[] = [
  'ciso',
  'compliance_manager',
  'security_engineer',
  'it_admin',
  'devops_engineer',
  'auditor',
  'vendor_manager',
];

function statusColor(status: string) {
  switch (status) {
    case 'active':
      return 'default';
    case 'invited':
      return 'secondary';
    case 'deactivated':
      return 'destructive';
    case 'locked':
      return 'destructive';
    default:
      return 'outline' as const;
  }
}

export default function UsersPage() {
  const { user: currentUser, canManageUsers } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showInvite, setShowInvite] = useState(false);
  const [inviteForm, setInviteForm] = useState({
    email: '',
    first_name: '',
    last_name: '',
    role: 'security_engineer' as GrcRole,
    password: '',
  });
  const [inviteError, setInviteError] = useState('');
  const [inviteLoading, setInviteLoading] = useState(false);
  const perPage = 20;

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage),
      });
      if (search) params.set('search', search);

      const res = await authFetch(`/api/v1/users?${params}`);
      const data = await res.json();

      if (res.ok) {
        setUsers(data.data || []);
        setTotal(data.meta?.total || 0);
      }
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  }, [page, search]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  async function handleInvite() {
    setInviteError('');
    setInviteLoading(true);
    try {
      const res = await authFetch('/api/v1/users', {
        method: 'POST',
        body: JSON.stringify(inviteForm),
      });
      const data = await res.json();
      if (res.ok) {
        setShowInvite(false);
        setInviteForm({
          email: '',
          first_name: '',
          last_name: '',
          role: 'security_engineer',
          password: '',
        });
        fetchUsers();
      } else {
        setInviteError(data.error?.message || 'Failed to create user');
      }
    } catch {
      setInviteError('Network error');
    } finally {
      setInviteLoading(false);
    }
  }

  async function handleDeactivate(userId: string) {
    if (!confirm('Are you sure you want to deactivate this user?')) return;
    try {
      const res = await authFetch(`/api/v1/users/${userId}/deactivate`, {
        method: 'POST',
      });
      if (res.ok) fetchUsers();
    } catch {
      // handle error
    }
  }

  async function handleReactivate(userId: string) {
    try {
      const res = await authFetch(`/api/v1/users/${userId}/reactivate`, {
        method: 'POST',
      });
      if (res.ok) fetchUsers();
    } catch {
      // handle error
    }
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Users className="h-8 w-8" />
            User Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage users and their GRC roles in your organization
          </p>
        </div>
        {canManageUsers && (
          <Button onClick={() => setShowInvite(!showInvite)}>
            <Plus className="h-4 w-4 mr-2" />
            Invite User
          </Button>
        )}
      </div>

      {/* Invite form */}
      {showInvite && canManageUsers && (
        <Card>
          <CardHeader>
            <CardTitle>Invite New User</CardTitle>
            <CardDescription>Add a team member to your organization</CardDescription>
          </CardHeader>
          <CardContent>
            {inviteError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive mb-4">
                {inviteError}
              </div>
            )}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              <div className="space-y-2">
                <Label>First Name</Label>
                <Input
                  value={inviteForm.first_name}
                  onChange={(e) => setInviteForm((f) => ({ ...f, first_name: e.target.value }))}
                  placeholder="First name"
                />
              </div>
              <div className="space-y-2">
                <Label>Last Name</Label>
                <Input
                  value={inviteForm.last_name}
                  onChange={(e) => setInviteForm((f) => ({ ...f, last_name: e.target.value }))}
                  placeholder="Last name"
                />
              </div>
              <div className="space-y-2">
                <Label>Email</Label>
                <Input
                  type="email"
                  value={inviteForm.email}
                  onChange={(e) => setInviteForm((f) => ({ ...f, email: e.target.value }))}
                  placeholder="user@company.com"
                />
              </div>
              <div className="space-y-2">
                <Label>Role</Label>
                <select
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  value={inviteForm.role}
                  onChange={(e) => setInviteForm((f) => ({ ...f, role: e.target.value as GrcRole }))}
                >
                  {ALL_ROLES.map((role) => (
                    <option key={role} value={role}>
                      {getRoleLabel(role)}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <Label>Temporary Password</Label>
                <Input
                  type="password"
                  value={inviteForm.password}
                  onChange={(e) => setInviteForm((f) => ({ ...f, password: e.target.value }))}
                  placeholder="Min 8 chars"
                />
              </div>
              <div className="flex items-end">
                <Button onClick={handleInvite} disabled={inviteLoading}>
                  {inviteLoading ? 'Creating...' : 'Create User'}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search users..."
            className="pl-10"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>
        <span className="text-sm text-muted-foreground">{total} users</span>
      </div>

      {/* User list */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              No users found
            </div>
          ) : (
            <div className="divide-y">
              {users.map((u) => {
                const initials =
                  `${u.first_name?.[0] || ''}${u.last_name?.[0] || ''}`.toUpperCase() ||
                  u.email[0].toUpperCase();
                const isSelf = u.id === currentUser?.id;

                return (
                  <div key={u.id} className="flex items-center gap-4 p-4 hover:bg-muted/50">
                    <Avatar className="h-10 w-10">
                      <AvatarFallback>{initials}</AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium truncate">
                          {u.first_name} {u.last_name}
                          {isSelf && (
                            <span className="text-muted-foreground ml-1">(you)</span>
                          )}
                        </p>
                        <Badge variant={statusColor(u.status)}>{u.status}</Badge>
                      </div>
                      <p className="text-xs text-muted-foreground truncate">{u.email}</p>
                    </div>
                    <Badge variant="outline">{getRoleLabel(u.role)}</Badge>
                    {canManageUsers && !isSelf && (
                      <div className="flex items-center gap-1">
                        {u.status === 'active' ? (
                          <Button
                            variant="ghost"
                            size="icon"
                            title="Deactivate"
                            onClick={() => handleDeactivate(u.id)}
                          >
                            <UserMinus className="h-4 w-4" />
                          </Button>
                        ) : u.status === 'deactivated' ? (
                          <Button
                            variant="ghost"
                            size="icon"
                            title="Reactivate"
                            onClick={() => handleReactivate(u.id)}
                          >
                            <UserCheck className="h-4 w-4" />
                          </Button>
                        ) : null}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
