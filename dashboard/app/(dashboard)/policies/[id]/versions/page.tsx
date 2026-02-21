'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import DOMPurify from 'dompurify';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import { ArrowLeft, GitCompare, History, ArrowRight } from 'lucide-react';
import {
  Policy, PolicyVersion,
  getPolicy, listPolicyVersions, comparePolicyVersions,
} from '@/lib/api';

export default function PolicyVersionHistoryPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [policy, setPolicy] = useState<Policy | null>(null);
  const [versions, setVersions] = useState<PolicyVersion[]>([]);
  const [loading, setLoading] = useState(true);

  // Comparison
  const [v1, setV1] = useState('');
  const [v2, setV2] = useState('');
  const [comparing, setComparing] = useState(false);
  const [compareData, setCompareData] = useState<{
    versions: PolicyVersion[];
    word_count_delta: number;
  } | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [polRes, versRes] = await Promise.all([
        getPolicy(id),
        listPolicyVersions(id, { per_page: '50' }),
      ]);
      setPolicy(polRes.data);
      setVersions(versRes.data);
      // Default to comparing last two versions
      if (versRes.data.length >= 2) {
        setV1(String(versRes.data[versRes.data.length - 1].version_number));
        setV2(String(versRes.data[0].version_number));
      }
    } catch (err) {
      console.error('Failed to fetch:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCompare = async () => {
    if (!v1 || !v2 || v1 === v2) return;
    try {
      setComparing(true);
      const res = await comparePolicyVersions(id, parseInt(v1), parseInt(v2));
      setCompareData(res.data);
    } catch (err) {
      console.error('Compare failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to compare versions');
    } finally {
      setComparing(false);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Loading...</div>;
  }
  if (!policy) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Policy not found</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push(`/policies/${id}`)}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="font-mono text-sm text-muted-foreground">{policy.identifier}</span>
          </div>
          <h1 className="text-2xl font-bold">Version History</h1>
          <p className="text-sm text-muted-foreground">{policy.title}</p>
        </div>
      </div>

      {/* Version List */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <History className="h-5 w-5" /> All Versions ({versions.length})
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Version</TableHead>
                <TableHead>Change Summary</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Words</TableHead>
                <TableHead>Author</TableHead>
                <TableHead>Sign-offs</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {versions.map((v) => (
                <TableRow key={v.id}>
                  <TableCell className="font-medium">
                    v{v.version_number}
                    {v.is_current && <Badge className="ml-2 text-xs" variant="secondary">Current</Badge>}
                  </TableCell>
                  <TableCell className="text-sm max-w-[300px] truncate">{v.change_summary || '—'}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs">{v.change_type || 'initial'}</Badge>
                  </TableCell>
                  <TableCell className="text-sm">{v.word_count || '—'}</TableCell>
                  <TableCell className="text-sm">{v.created_by?.name || '—'}</TableCell>
                  <TableCell className="text-sm">
                    {v.signoff_summary ? (
                      <span>
                        {v.signoff_summary.approved}/{v.signoff_summary.total}
                        {v.signoff_summary.pending > 0 && (
                          <span className="text-yellow-600 ml-1">({v.signoff_summary.pending} pending)</span>
                        )}
                      </span>
                    ) : '—'}
                  </TableCell>
                  <TableCell className="text-sm">{v.created_at ? new Date(v.created_at).toLocaleDateString() : '—'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Compare Section */}
      {versions.length >= 2 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <GitCompare className="h-5 w-5" /> Compare Versions
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-end gap-4">
              <div className="w-[150px]">
                <label className="text-sm font-medium mb-1 block">From (older)</label>
                <Select value={v1} onValueChange={setV1}>
                  <SelectTrigger><SelectValue placeholder="Version" /></SelectTrigger>
                  <SelectContent>
                    {versions.map(v => (
                      <SelectItem key={v.id} value={String(v.version_number)}>
                        v{v.version_number}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <ArrowRight className="h-5 w-5 text-muted-foreground mb-2" />
              <div className="w-[150px]">
                <label className="text-sm font-medium mb-1 block">To (newer)</label>
                <Select value={v2} onValueChange={setV2}>
                  <SelectTrigger><SelectValue placeholder="Version" /></SelectTrigger>
                  <SelectContent>
                    {versions.map(v => (
                      <SelectItem key={v.id} value={String(v.version_number)}>
                        v{v.version_number}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <Button onClick={handleCompare} disabled={comparing || !v1 || !v2 || v1 === v2}>
                {comparing ? 'Comparing...' : 'Compare'}
              </Button>
            </div>

            {compareData && (
              <div className="space-y-4 mt-4">
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-muted-foreground">
                    Word count change: <strong className={compareData.word_count_delta >= 0 ? 'text-green-600' : 'text-red-600'}>
                      {compareData.word_count_delta >= 0 ? '+' : ''}{compareData.word_count_delta}
                    </strong>
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  {compareData.versions.map((ver, idx) => (
                    <div key={ver.version_number}>
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="font-medium">v{ver.version_number}</h4>
                        <div className="text-xs text-muted-foreground">
                          {ver.word_count} words • {ver.created_by?.name} • {ver.created_at ? new Date(ver.created_at).toLocaleDateString() : ''}
                        </div>
                      </div>
                      <div className="text-xs text-muted-foreground mb-2">
                        {ver.change_summary}
                      </div>
                      <div
                        className={`prose dark:prose-invert prose-sm max-w-none p-4 border rounded-md max-h-[500px] overflow-y-auto ${
                          idx === 0 ? 'bg-red-50/50 dark:bg-red-950/10' : 'bg-green-50/50 dark:bg-green-950/10'
                        }`}
                        dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(ver.content || '<p>(No content)</p>') }}
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
