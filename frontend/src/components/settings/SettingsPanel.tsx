'use client'

import { useState, useEffect } from 'react'
import { useAuthStore } from '@/lib/store/authStore'
import apiClient from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

interface DailyStat {
  date: string
  total_generations: number
  successful_generations: number
  total_credits_used: number
}

interface ApiConfig {
  id: number
  config_name: string
  api_base_url: string
  model: string
  is_active: boolean
}

export default function SettingsPanel() {
  const { user } = useAuthStore()
  const [stats, setStats] = useState<DailyStat[]>([])
  const [configs, setConfigs] = useState<ApiConfig[]>([])
  const [showConfigForm, setShowConfigForm] = useState(false)
  const [configName, setConfigName] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [apiBaseUrl, setApiBaseUrl] = useState('')
  const [model, setModel] = useState('gpt-image-2')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    loadStats()
    if (user?.role === 'admin') {
      loadConfigs()
    }
  }, [user])

  const loadStats = async () => {
    try {
      const res: any = await apiClient.get('/stats/daily', { params: { days: 30 } })
      setStats(res.data || [])
    } catch (err) {
      console.error('Failed to load stats:', err)
    }
  }

  const loadConfigs = async () => {
    try {
      const res: any = await apiClient.get('/config')
      setConfigs(res.data || [])
    } catch (err) {
      console.error('Failed to load configs:', err)
    }
  }

  const handleSaveConfig = async () => {
    if (!configName || !apiKey) {
      setError('Config name and API key are required')
      return
    }
    setSaving(true)
    setError('')
    try {
      await apiClient.post('/config', {
        config_name: configName,
        api_key: apiKey,
        api_base_url: apiBaseUrl,
        model,
        is_active: true,
      })
      setShowConfigForm(false)
      setConfigName('')
      setApiKey('')
      setApiBaseUrl('')
      loadConfigs()
    } catch (err: any) {
      setError(err?.message || 'Failed to save config')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteConfig = async (id: number) => {
    if (!confirm('Delete this API config?')) return
    try {
      await apiClient.delete(`/config/${id}`)
      loadConfigs()
    } catch (err) {
      console.error('Failed to delete config:', err)
    }
  }

  const totalGenerations = stats.reduce((sum, s) => sum + s.total_generations, 0)
  const totalCredits = stats.reduce((sum, s) => sum + s.total_credits_used, 0)
  const totalSuccess = stats.reduce((sum, s) => sum + s.successful_generations, 0)

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-foreground mb-6">Settings</h1>

      <Tabs defaultValue="stats">
        <TabsList>
          <TabsTrigger value="stats">Usage Stats</TabsTrigger>
          {user?.role === 'admin' && (
            <TabsTrigger value="api">API Config</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="stats" className="space-y-6">
          <div className="grid grid-cols-3 gap-4">
            <Card>
              <CardContent className="p-4">
                <div className="text-3xl font-bold text-foreground">{totalGenerations}</div>
                <div className="text-sm text-muted-foreground mt-1">Total Generations</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-3xl font-bold text-emerald-400">{totalSuccess}</div>
                <div className="text-sm text-muted-foreground mt-1">Successful</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-3xl font-bold text-yellow-400">{totalCredits}</div>
                <div className="text-sm text-muted-foreground mt-1">Credits Used</div>
              </CardContent>
            </Card>
          </div>

          <div>
            <h3 className="text-lg font-semibold text-foreground mb-4">Daily Usage (Last 30 Days)</h3>
            {stats.length === 0 ? (
              <p className="text-muted-foreground">No usage data yet</p>
            ) : (
              <Card>
                <CardContent className="p-0">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-border">
                        <th className="text-left p-3 text-sm text-muted-foreground">Date</th>
                        <th className="text-right p-3 text-sm text-muted-foreground">Generations</th>
                        <th className="text-right p-3 text-sm text-muted-foreground">Success</th>
                        <th className="text-right p-3 text-sm text-muted-foreground">Credits</th>
                      </tr>
                    </thead>
                    <tbody>
                      {stats.map((stat) => (
                        <tr key={stat.date} className="border-b border-border last:border-0">
                          <td className="p-3 text-sm text-foreground">{stat.date}</td>
                          <td className="p-3 text-sm text-foreground text-right">{stat.total_generations}</td>
                          <td className="p-3 text-sm text-emerald-400 text-right">{stat.successful_generations}</td>
                          <td className="p-3 text-sm text-yellow-400 text-right">{stat.total_credits_used}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        <TabsContent value="api" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-semibold text-foreground">API Configurations</h3>
            <Button size="sm" onClick={() => setShowConfigForm(!showConfigForm)}>
              {showConfigForm ? 'Cancel' : '+ Add Config'}
            </Button>
          </div>

          {showConfigForm && (
            <Card>
              <CardContent className="p-4 space-y-3">
                {error && (
                  <div className="bg-destructive/10 border border-destructive/50 rounded-lg p-3 text-destructive text-sm">
                    {error}
                  </div>
                )}
                <div className="space-y-2">
                  <Label>Config Name</Label>
                  <Input
                    value={configName}
                    onChange={(e) => setConfigName(e.target.value)}
                    placeholder="e.g. My API Key"
                  />
                </div>
                <div className="space-y-2">
                  <Label>API Key</Label>
                  <Input
                    type="password"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    placeholder="sk-..."
                  />
                </div>
                <div className="space-y-2">
                  <Label>API Base URL (for proxy)</Label>
                  <Input
                    value={apiBaseUrl}
                    onChange={(e) => setApiBaseUrl(e.target.value)}
                    placeholder="https://api.openai.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Model</Label>
                  <Input
                    value={model}
                    onChange={(e) => setModel(e.target.value)}
                    placeholder="gpt-image-2"
                  />
                </div>
                <Button onClick={handleSaveConfig} disabled={saving} size="sm">
                  {saving ? 'Saving...' : 'Save Config'}
                </Button>
              </CardContent>
            </Card>
          )}

          {configs.length === 0 ? (
            <p className="text-muted-foreground">No API configurations yet</p>
          ) : (
            <div className="space-y-3">
              {configs.map((config) => (
                <Card key={config.id}>
                  <CardContent className="p-4 flex justify-between items-center">
                    <div>
                      <div className="text-foreground font-medium">{config.config_name}</div>
                      <div className="text-sm text-muted-foreground mt-1">
                        {config.api_base_url || 'Default API'} · {config.model}
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <Badge variant={config.is_active ? 'success' : 'secondary'}>
                        {config.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteConfig(config.id)}
                        className="text-muted-foreground hover:text-destructive"
                      >
                        Delete
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
