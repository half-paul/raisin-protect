'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/lib/auth-context';
import { authFetch } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Building2, Save, Key, Shield } from 'lucide-react';

interface Organization {
  id: string;
  name: string;
  slug: string;
  domain: string;
  status: string;
  settings: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export default function SettingsPage() {
  const { user, isAdmin } = useAuth();
  const [org, setOrg] = useState<Organization | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState('');
  const [error, setError] = useState('');
  const [form, setForm] = useState({ name: '', domain: '' });

  // Password change
  const [passwordForm, setPasswordForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  useEffect(() => {
    async function fetchOrg() {
      try {
        const res = await authFetch('/api/v1/organizations/current');
        const data = await res.json();
        if (res.ok && data.data) {
          setOrg(data.data);
          setForm({ name: data.data.name || '', domain: data.data.domain || '' });
        }
      } catch {
        // handle error
      } finally {
        setLoading(false);
      }
    }
    fetchOrg();
  }, []);

  async function handleSaveOrg() {
    setError('');
    setSuccess('');
    setSaving(true);
    try {
      const res = await authFetch('/api/v1/organizations/current', {
        method: 'PUT',
        body: JSON.stringify(form),
      });
      const data = await res.json();
      if (res.ok) {
        setOrg(data.data);
        setSuccess('Organization updated successfully');
      } else {
        setError(data.error?.message || 'Failed to update organization');
      }
    } catch {
      setError('Network error');
    } finally {
      setSaving(false);
    }
  }

  async function handleChangePassword() {
    setPasswordError('');
    setPasswordSuccess('');

    if (passwordForm.new_password !== passwordForm.confirm_password) {
      setPasswordError('Passwords do not match');
      return;
    }

    setPasswordLoading(true);
    try {
      const res = await authFetch('/api/v1/auth/change-password', {
        method: 'POST',
        body: JSON.stringify({
          current_password: passwordForm.current_password,
          new_password: passwordForm.new_password,
        }),
      });
      const data = await res.json();
      if (res.ok) {
        setPasswordSuccess('Password changed successfully. Please log in again.');
        setPasswordForm({ current_password: '', new_password: '', confirm_password: '' });
      } else {
        setPasswordError(data.error?.message || 'Failed to change password');
      }
    } catch {
      setPasswordError('Network error');
    } finally {
      setPasswordLoading(false);
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <Building2 className="h-8 w-8" />
          Organization Settings
        </h1>
        <p className="text-muted-foreground mt-1">Manage your organization and account settings</p>
      </div>

      {/* Organization info */}
      <Card>
        <CardHeader>
          <CardTitle>Organization Details</CardTitle>
          <CardDescription>
            Update your organization name and domain
            {!isAdmin && ' (read-only for your role)'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
          )}
          {success && (
            <div className="rounded-md bg-green-500/10 p-3 text-sm text-green-700 dark:text-green-400">
              {success}
            </div>
          )}
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : (
            <>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>Organization Name</Label>
                  <Input
                    value={form.name}
                    onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                    disabled={!isAdmin}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Domain</Label>
                  <Input
                    value={form.domain}
                    onChange={(e) => setForm((f) => ({ ...f, domain: e.target.value }))}
                    placeholder="acme.example.com"
                    disabled={!isAdmin}
                  />
                </div>
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>Slug (read-only)</Label>
                  <Input value={org?.slug || ''} disabled />
                </div>
                <div className="space-y-2">
                  <Label>Status</Label>
                  <div className="pt-2">
                    <Badge variant={org?.status === 'active' ? 'default' : 'secondary'}>
                      {org?.status || 'unknown'}
                    </Badge>
                  </div>
                </div>
              </div>
            </>
          )}
        </CardContent>
        {isAdmin && (
          <CardFooter>
            <Button onClick={handleSaveOrg} disabled={saving}>
              <Save className="h-4 w-4 mr-2" />
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </CardFooter>
        )}
      </Card>

      <Separator />

      {/* Change password */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            Change Password
          </CardTitle>
          <CardDescription>Update your account password</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {passwordError && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {passwordError}
            </div>
          )}
          {passwordSuccess && (
            <div className="rounded-md bg-green-500/10 p-3 text-sm text-green-700 dark:text-green-400">
              {passwordSuccess}
            </div>
          )}
          <div className="max-w-md space-y-4">
            <div className="space-y-2">
              <Label>Current Password</Label>
              <Input
                type="password"
                value={passwordForm.current_password}
                onChange={(e) =>
                  setPasswordForm((f) => ({ ...f, current_password: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label>New Password</Label>
              <Input
                type="password"
                value={passwordForm.new_password}
                onChange={(e) =>
                  setPasswordForm((f) => ({ ...f, new_password: e.target.value }))
                }
                placeholder="Min 8 chars, upper/lower/number/special"
              />
            </div>
            <div className="space-y-2">
              <Label>Confirm New Password</Label>
              <Input
                type="password"
                value={passwordForm.confirm_password}
                onChange={(e) =>
                  setPasswordForm((f) => ({ ...f, confirm_password: e.target.value }))
                }
              />
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <Button onClick={handleChangePassword} disabled={passwordLoading} variant="outline">
            <Shield className="h-4 w-4 mr-2" />
            {passwordLoading ? 'Changing...' : 'Change Password'}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
