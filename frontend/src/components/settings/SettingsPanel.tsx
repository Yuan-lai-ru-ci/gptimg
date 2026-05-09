'use client'

import { useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '@/lib/store/authStore'
import apiClient from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import FormInput from '@/components/ui/form-input'
const Input = FormInput

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
  max_requests_per_minute: number
}

interface AdminOverview {
  total_users: number
  active_users: number
  suspended_users: number
  total_generations: number
  successful_runs: number
  failed_runs: number
  total_credits_used: number
  active_api_configs: number
  healthy_api_configs: number
}

interface AdminUserSummary {
  id: number
  username: string
  email: string
  credits: number
  quota_limit: number
  role: string
  status: string
  total_generations: number
  successful_generations: number
  total_credits_used: number
  created_at: string
}

interface ApiPoolStatus {
  id: number
  config_name: string
  api_base_url: string
  model: string
  is_active: boolean
  max_requests_per_minute: number
  healthy: boolean
  status_code: number
  status_message: string
  checked_at: string
}

interface LLMConfig {
  id: number
  config_name: string
  api_base_url: string
  model: string
  is_active: boolean
  max_requests_per_minute: number
}

interface LLMPoolStatus {
  id: number
  config_name: string
  api_base_url: string
  model: string
  is_active: boolean
  max_requests_per_minute: number
  healthy: boolean
  status_code: number
  status_message: string
  checked_at: string
}

type AdminTab = 'overview' | 'users' | 'api' | 'llm'

export default function SettingsPanel() {
  const { user } = useAuthStore()
  const isAdmin = user?.role === 'admin'

  const [activeTab, setActiveTab] = useState<AdminTab>('overview')
  const [stats, setStats] = useState<DailyStat[]>([])
  const [overview, setOverview] = useState<AdminOverview | null>(null)
  const [users, setUsers] = useState<AdminUserSummary[]>([])
  const [configs, setConfigs] = useState<ApiConfig[]>([])
  const [poolStatus, setPoolStatus] = useState<ApiPoolStatus[]>([])
  const [llmConfigs, setLlmConfigs] = useState<LLMConfig[]>([])
  const [llmPoolStatus, setLlmPoolStatus] = useState<LLMPoolStatus[]>([])

  const [showConfigForm, setShowConfigForm] = useState(false)
  const [showLLMConfigForm, setShowLLMConfigForm] = useState(false)
  const [showUserForm, setShowUserForm] = useState(false)
  const [savingConfig, setSavingConfig] = useState(false)
  const [savingLLMConfig, setSavingLLMConfig] = useState(false)
  const [savingUser, setSavingUser] = useState(false)
  const [error, setError] = useState('')

  const [configName, setConfigName] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [apiBaseUrl, setApiBaseUrl] = useState('')
  const [model, setModel] = useState('gpt-image-2')
  const [maxRequestsPerMinute, setMaxRequestsPerMinute] = useState('5')
  const [isActive, setIsActive] = useState(true)
  const [llmConfigName, setLlmConfigName] = useState('')
  const [llmApiKey, setLlmApiKey] = useState('')
  const [llmApiBaseUrl, setLlmApiBaseUrl] = useState('https://api.deepseek.com')
  const [llmModel, setLlmModel] = useState('deepseek-chat')
  const [llmMaxRequestsPerMinute, setLlmMaxRequestsPerMinute] = useState('5')
  const [llmIsActive, setLlmIsActive] = useState(true)

  const [newUsername, setNewUsername] = useState('')
  const [newEmail, setNewEmail] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [newQuota, setNewQuota] = useState('50')
  const [newRole, setNewRole] = useState('user')

  useEffect(() => {
    if (isAdmin) {
      void loadAdminData()
    } else {
      void loadUserStats()
    }
  }, [isAdmin])

  const loadUserStats = async () => {
    try {
      const res: any = await apiClient.get('/stats/daily', { params: { days: 30 } })
      setStats(res.data || [])
    } catch (err) {
      console.error('Failed to load stats:', err)
    }
  }

  const loadAdminData = async () => {
    try {
      const [overviewRes, usersRes, configsRes, poolRes, llmConfigsRes, llmPoolRes] = await Promise.all([
        apiClient.get('/admin/overview'),
        apiClient.get('/admin/users'),
        apiClient.get('/config'),
        apiClient.get('/admin/api-pool'),
        apiClient.get('/llm-config'),
        apiClient.get('/admin/llm-pool'),
      ])
      setOverview(overviewRes.data)
      setUsers(usersRes.data || [])
      setConfigs(configsRes.data || [])
      setPoolStatus(poolRes.data || [])
      setLlmConfigs(llmConfigsRes.data || [])
      setLlmPoolStatus(llmPoolRes.data || [])
    } catch (err) {
      console.error('Failed to load admin data:', err)
    }
  }

  const handleCreateUser = async () => {
    if (!newUsername || !newEmail || !newPassword) {
      setError('Username, email and password are required')
      return
    }

    setSavingUser(true)
    setError('')
    try {
      await apiClient.post('/admin/users', {
        username: newUsername,
        email: newEmail,
        password: newPassword,
        quota_limit: Number(newQuota) || 50,
        role: newRole,
      })
      setNewUsername('')
      setNewEmail('')
      setNewPassword('')
      setNewQuota('50')
      setNewRole('user')
      setShowUserForm(false)
      await loadAdminData()
    } catch (err: any) {
      setError(err?.message || 'Failed to create user')
    } finally {
      setSavingUser(false)
    }
  }

  const handleSaveConfig = async () => {
    if (!configName || !apiKey) {
      setError('Config name and API key are required')
      return
    }

    setSavingConfig(true)
    setError('')
    try {
      await apiClient.post('/config', {
        config_name: configName,
        api_key: apiKey,
        api_base_url: apiBaseUrl,
        model,
        is_active: isActive,
        max_requests_per_minute: Number(maxRequestsPerMinute) || 5,
      })
      setConfigName('')
      setApiKey('')
      setApiBaseUrl('')
      setModel('gpt-image-2')
      setMaxRequestsPerMinute('5')
      setIsActive(true)
      setShowConfigForm(false)
      await loadAdminData()
    } catch (err: any) {
      setError(err?.message || 'Failed to save config')
    } finally {
      setSavingConfig(false)
    }
  }

  const handleSaveLLMConfig = async () => {
    if (!llmConfigName || !llmApiKey) {
      setError('LLM config name and API key are required')
      return
    }

    setSavingLLMConfig(true)
    setError('')
    try {
      await apiClient.post('/llm-config', {
        config_name: llmConfigName,
        api_key: llmApiKey,
        api_base_url: llmApiBaseUrl,
        model: llmModel,
        is_active: llmIsActive,
        max_requests_per_minute: Number(llmMaxRequestsPerMinute) || 5,
      })
      setLlmConfigName('')
      setLlmApiKey('')
      setLlmApiBaseUrl('https://api.deepseek.com')
      setLlmModel('deepseek-chat')
      setLlmMaxRequestsPerMinute('5')
      setLlmIsActive(true)
      setShowLLMConfigForm(false)
      await loadAdminData()
    } catch (err: any) {
      setError(err?.message || 'Failed to save LLM config')
    } finally {
      setSavingLLMConfig(false)
    }
  }

  const handleToggleConfig = async (config: ApiConfig) => {
    try {
      await apiClient.put(`/config/${config.id}`, {
        config_name: config.config_name,
        api_base_url: config.api_base_url,
        model: config.model,
        is_active: !config.is_active,
        max_requests_per_minute: config.max_requests_per_minute,
      })
      await loadAdminData()
    } catch (err) {
      console.error('Failed to update config:', err)
    }
  }

  const handleDeleteConfig = async (id: number) => {
    if (!confirm('Delete this API config?')) return
    try {
      await apiClient.delete(`/config/${id}`)
      await loadAdminData()
    } catch (err) {
      console.error('Failed to delete config:', err)
    }
  }

  const handleToggleLLMConfig = async (config: LLMConfig) => {
    try {
      await apiClient.put(`/llm-config/${config.id}`, {
        config_name: config.config_name,
        api_base_url: config.api_base_url,
        model: config.model,
        is_active: !config.is_active,
        max_requests_per_minute: config.max_requests_per_minute,
      })
      await loadAdminData()
    } catch (err) {
      console.error('Failed to update LLM config:', err)
    }
  }

  const handleDeleteLLMConfig = async (id: number) => {
    if (!confirm('Delete this LLM config?')) return
    try {
      await apiClient.delete(`/llm-config/${id}`)
      await loadAdminData()
    } catch (err) {
      console.error('Failed to delete LLM config:', err)
    }
  }

  const handleDeleteUser = async (account: AdminUserSummary) => {
    if (!confirm(`Delete account ${account.username}? This will remove its sessions, images and usage data.`)) {
      return
    }

    try {
      await apiClient.delete(`/admin/users/${account.id}`)
      await loadAdminData()
    } catch (err) {
      console.error('Failed to delete user:', err)
    }
  }

  const totalGenerations = stats.reduce((sum, s) => sum + s.total_generations, 0)
  const totalCredits = stats.reduce((sum, s) => sum + s.total_credits_used, 0)
  const totalSuccess = stats.reduce((sum, s) => sum + s.successful_generations, 0)

  const sortedUsers = useMemo(
    () =>
      [...users].sort((a, b) => {
        if (a.role === b.role) return a.username.localeCompare(b.username)
        return a.role === 'admin' ? -1 : 1
      }),
    [users]
  )

  if (!isAdmin) {
    return (
      <div className="mx-auto max-w-4xl p-4 sm:p-6">
        <h1 className="mb-6 text-2xl font-bold text-foreground">Usage Stats</h1>
        <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="text-3xl font-bold text-foreground">{totalGenerations}</div>
            <div className="mt-1 text-sm text-muted-foreground">Total Generations</div>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="text-3xl font-bold text-emerald-400">{totalSuccess}</div>
            <div className="mt-1 text-sm text-muted-foreground">Successful</div>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="text-3xl font-bold text-yellow-400">{totalCredits}</div>
            <div className="mt-1 text-sm text-muted-foreground">Credits Used</div>
          </div>
        </div>

        <h3 className="mb-4 text-lg font-semibold text-foreground">Daily Usage (Last 30 Days)</h3>
        {stats.length === 0 ? (
          <p className="text-muted-foreground">No usage data yet</p>
        ) : (
          <div className="overflow-x-auto rounded-xl border border-border bg-card">
            <table className="min-w-[640px] w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="p-3 text-left text-sm text-muted-foreground">Date</th>
                  <th className="p-3 text-right text-sm text-muted-foreground">Generations</th>
                  <th className="p-3 text-right text-sm text-muted-foreground">Success</th>
                  <th className="p-3 text-right text-sm text-muted-foreground">Credits</th>
                </tr>
              </thead>
              <tbody>
                {stats.map((stat) => (
                  <tr key={stat.date} className="border-b border-border last:border-0">
                    <td className="p-3 text-sm text-foreground">{stat.date}</td>
                    <td className="p-3 text-right text-sm text-foreground">{stat.total_generations}</td>
                    <td className="p-3 text-right text-sm text-emerald-400">{stat.successful_generations}</td>
                    <td className="p-3 text-right text-sm text-yellow-400">{stat.total_credits_used}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-6xl p-4 sm:p-6">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Admin Console</h1>

      {error && (
        <div className="mb-4 rounded-lg border border-red-800 bg-red-900/20 p-3 text-sm text-red-400">
          {error}
        </div>
      )}

      <div className="mb-6 flex flex-wrap gap-3">
        {(['overview', 'users', 'api', 'llm'] as AdminTab[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`rounded-lg px-4 py-2 text-sm ${
              activeTab === tab ? 'bg-white text-black' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {tab === 'overview' ? 'Overview' : tab === 'users' ? 'Users' : tab === 'api' ? 'Image API' : 'LLM Pool'}
          </button>
        ))}
      </div>

      {activeTab === 'overview' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3 xl:grid-cols-5">
            <MetricCard label="Total Users" value={overview?.total_users ?? 0} />
            <MetricCard label="Active Users" value={overview?.active_users ?? 0} accent="emerald" />
            <MetricCard label="Suspended" value={overview?.suspended_users ?? 0} accent="rose" />
            <MetricCard label="Generations" value={overview?.total_generations ?? 0} />
            <MetricCard label="Credits Used" value={overview?.total_credits_used ?? 0} accent="yellow" />
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="rounded-xl border border-border bg-card p-4">
              <div className="text-sm text-muted-foreground">API Pool Health</div>
              <div className="mt-2 text-2xl font-bold text-foreground">
                {overview?.healthy_api_configs ?? 0} / {overview?.active_api_configs ?? 0}
              </div>
              <div className="mt-1 text-sm text-muted-foreground">Healthy active configs</div>
            </div>
            <div className="rounded-xl border border-border bg-card p-4">
              <div className="text-sm text-muted-foreground">Success Runs</div>
              <div className="mt-2 text-2xl font-bold text-emerald-400">{overview?.successful_runs ?? 0}</div>
            </div>
            <div className="rounded-xl border border-border bg-card p-4">
              <div className="text-sm text-muted-foreground">Failed Runs</div>
              <div className="mt-2 text-2xl font-bold text-rose-400">{overview?.failed_runs ?? 0}</div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'users' && (
        <div className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h2 className="text-lg font-semibold text-foreground">User Quota Tracking</h2>
            <Button size="sm" onClick={() => setShowUserForm((prev) => !prev)}>
              {showUserForm ? 'Cancel' : '+ Issue Account'}
            </Button>
          </div>

          {showUserForm && (
            <div className="grid grid-cols-1 gap-3 rounded-xl border border-border bg-card p-4 md:grid-cols-2">
              <Input label="Username" value={newUsername} onChange={(e) => setNewUsername(e.target.value)} />
              <Input label="Email" type="email" value={newEmail} onChange={(e) => setNewEmail(e.target.value)} />
              <Input label="Password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
              <Input label="Quota" type="number" value={newQuota} onChange={(e) => setNewQuota(e.target.value)} />
              <div>
                <label className="mb-2 block text-sm font-medium text-secondary-foreground">Role</label>
                <select
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value)}
                  className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground"
                >
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              <div className="flex items-end">
                <Button onClick={handleCreateUser} disabled={savingUser} className="w-full">
                  {savingUser ? 'Creating...' : 'Create Account'}
                </Button>
              </div>
            </div>
          )}

          <div className="space-y-3">
            {sortedUsers.map((account) => {
              const remaining = Math.min(account.credits, account.quota_limit)
              const used = Math.max(account.quota_limit - remaining, 0)
              const progress = account.quota_limit > 0 ? Math.min((used / account.quota_limit) * 100, 100) : 0
              return (
                <div key={account.id} className="rounded-xl border border-border bg-card p-4">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-foreground font-medium">{account.username}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${account.role === 'admin' ? 'bg-blue-900/30 text-blue-300' : 'bg-gray-800 text-muted-foreground'}`}>{account.role}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${account.status === 'active' ? 'bg-emerald-900/30 text-emerald-400' : 'bg-rose-900/30 text-rose-300'}`}>{account.status}</span>
                      </div>
                      <div className="truncate text-sm text-muted-foreground">{account.email}</div>
                    </div>
                    <div className="grid grid-cols-2 gap-4 text-sm text-secondary-foreground sm:grid-cols-4">
                      <div><div className="text-muted-foreground">Remaining</div><div className="text-foreground">{remaining}</div></div>
                      <div><div className="text-muted-foreground">Quota</div><div className="text-foreground">{account.quota_limit}</div></div>
                      <div><div className="text-muted-foreground">Generated</div><div className="text-foreground">{account.total_generations}</div></div>
                      <div><div className="text-muted-foreground">Credits Used</div><div className="text-foreground">{account.total_credits_used}</div></div>
                    </div>
                  </div>
                  <div className="mt-4">
                    <div className="mb-2 flex items-center justify-between text-xs text-muted-foreground">
                      <span>Usage progress</span>
                      <span>{used}/{account.quota_limit}</span>
                    </div>
                    <div className="h-2 overflow-hidden rounded-full bg-background">
                      <div className="h-full rounded-full bg-white" style={{ width: `${progress}%` }} />
                    </div>
                  </div>
                  {account.role !== 'admin' && (
                    <div className="mt-4 flex justify-end">
                      <Button size="sm" variant="ghost" onClick={() => void handleDeleteUser(account)}>
                        Delete Account
                      </Button>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {activeTab === 'api' && (
        <div className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h2 className="text-lg font-semibold text-foreground">API Pool</h2>
            <div className="flex gap-2">
              <Button size="sm" variant="secondary" onClick={() => void loadAdminData()}>
                Refresh Status
              </Button>
              <Button size="sm" onClick={() => setShowConfigForm((prev) => !prev)}>
                {showConfigForm ? 'Cancel' : '+ Add Config'}
              </Button>
            </div>
          </div>

          {showConfigForm && (
            <div className="grid grid-cols-1 gap-3 rounded-xl border border-border bg-card p-4 md:grid-cols-2">
              <Input label="Config Name" value={configName} onChange={(e) => setConfigName(e.target.value)} />
              <Input label="API Key" type="password" value={apiKey} onChange={(e) => setApiKey(e.target.value)} placeholder="sk-..." />
              <Input label="API Base URL" value={apiBaseUrl} onChange={(e) => setApiBaseUrl(e.target.value)} placeholder="https://api.openai.com" />
              <Input label="Model" value={model} onChange={(e) => setModel(e.target.value)} placeholder="gpt-image-2" />
              <Input label="Max Requests Per Minute" type="number" value={maxRequestsPerMinute} onChange={(e) => setMaxRequestsPerMinute(e.target.value)} />
              <div className="flex items-end gap-3">
                <label className="flex items-center gap-2 text-sm text-secondary-foreground">
                  <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
                  Active in pool
                </label>
                <Button onClick={handleSaveConfig} disabled={savingConfig} className="ml-auto">
                  {savingConfig ? 'Saving...' : 'Save Config'}
                </Button>
              </div>
            </div>
          )}

          <div className="space-y-3">
            {poolStatus.map((status) => {
              const config = configs.find((item) => item.id === status.id)
              return (
                <div key={status.id} className="rounded-xl border border-border bg-card p-4">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-foreground">{status.config_name}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${status.is_active ? 'bg-emerald-900/30 text-emerald-400' : 'bg-gray-800 text-muted-foreground'}`}>{status.is_active ? 'Active' : 'Inactive'}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${status.healthy ? 'bg-blue-900/30 text-blue-300' : 'bg-rose-900/30 text-rose-300'}`}>{status.healthy ? 'Connected' : 'Unavailable'}</span>
                      </div>
                      <div className="mt-1 truncate text-sm text-muted-foreground">
                        {(status.api_base_url || 'Default API')} | {status.model}
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {status.status_code ? `HTTP ${status.status_code}` : 'No response code'} | {status.status_message}
                      </div>
                    </div>
                    {config ? (
                      <div className="flex flex-wrap gap-2">
                        <Button size="sm" variant="secondary" onClick={() => void handleToggleConfig(config)}>
                          {config.is_active ? 'Disable' : 'Enable'}
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => void handleDeleteConfig(config.id)}>
                          Delete
                        </Button>
                      </div>
                    ) : null}
                  </div>
                </div>
              )
            })}

            {poolStatus.length === 0 && <p className="text-muted-foreground">No API configurations yet</p>}
          </div>
        </div>
      )}

      {activeTab === 'llm' && (
        <div className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h2 className="text-lg font-semibold text-foreground">LLM Pool</h2>
            <div className="flex gap-2">
              <Button size="sm" variant="secondary" onClick={() => void loadAdminData()}>
                Refresh Status
              </Button>
              <Button size="sm" onClick={() => setShowLLMConfigForm((prev) => !prev)}>
                {showLLMConfigForm ? 'Cancel' : '+ Add LLM Config'}
              </Button>
            </div>
          </div>

          {showLLMConfigForm && (
            <div className="grid grid-cols-1 gap-3 rounded-xl border border-border bg-card p-4 md:grid-cols-2">
              <Input label="Config Name" value={llmConfigName} onChange={(e) => setLlmConfigName(e.target.value)} />
              <Input label="API Key" type="password" value={llmApiKey} onChange={(e) => setLlmApiKey(e.target.value)} placeholder="DeepSeek key" />
              <Input label="API Base URL" value={llmApiBaseUrl} onChange={(e) => setLlmApiBaseUrl(e.target.value)} placeholder="https://api.deepseek.com" />
              <Input label="Model" value={llmModel} onChange={(e) => setLlmModel(e.target.value)} placeholder="deepseek-chat" />
              <Input label="Max Requests Per Minute" type="number" value={llmMaxRequestsPerMinute} onChange={(e) => setLlmMaxRequestsPerMinute(e.target.value)} />
              <div className="flex items-end gap-3">
                <label className="flex items-center gap-2 text-sm text-secondary-foreground">
                  <input type="checkbox" checked={llmIsActive} onChange={(e) => setLlmIsActive(e.target.checked)} />
                  Active in pool
                </label>
                <Button onClick={handleSaveLLMConfig} disabled={savingLLMConfig} className="ml-auto">
                  {savingLLMConfig ? 'Saving...' : 'Save LLM Config'}
                </Button>
              </div>
            </div>
          )}

          <div className="space-y-3">
            {llmPoolStatus.map((status) => {
              const config = llmConfigs.find((item) => item.id === status.id)
              return (
                <div key={status.id} className="rounded-xl border border-border bg-card p-4">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-foreground">{status.config_name}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${status.is_active ? 'bg-emerald-900/30 text-emerald-400' : 'bg-gray-800 text-muted-foreground'}`}>{status.is_active ? 'Active' : 'Inactive'}</span>
                        <span className={`rounded px-2 py-0.5 text-xs ${status.healthy ? 'bg-blue-900/30 text-blue-300' : 'bg-rose-900/30 text-rose-300'}`}>{status.healthy ? 'Connected' : 'Unavailable'}</span>
                      </div>
                      <div className="mt-1 truncate text-sm text-muted-foreground">
                        {(status.api_base_url || 'Default LLM API')} | {status.model}
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {status.status_code ? `HTTP ${status.status_code}` : 'No response code'} | {status.status_message}
                      </div>
                    </div>
                    {config ? (
                      <div className="flex flex-wrap gap-2">
                        <Button size="sm" variant="secondary" onClick={() => void handleToggleLLMConfig(config)}>
                          {config.is_active ? 'Disable' : 'Enable'}
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => void handleDeleteLLMConfig(config.id)}>
                          Delete
                        </Button>
                      </div>
                    ) : null}
                  </div>
                </div>
              )
            })}

            {llmPoolStatus.length === 0 && <p className="text-muted-foreground">No LLM configurations yet</p>}
          </div>
        </div>
      )}
    </div>
  )
}

function MetricCard({
  label,
  value,
  accent = 'white',
}: {
  label: string
  value: number
  accent?: 'white' | 'emerald' | 'yellow' | 'rose'
}) {
  const accentClass =
    accent === 'emerald'
      ? 'text-emerald-400'
      : accent === 'yellow'
        ? 'text-yellow-400'
        : accent === 'rose'
          ? 'text-rose-400'
          : 'text-foreground'

  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className={`text-3xl font-bold ${accentClass}`}>{value}</div>
      <div className="mt-1 text-sm text-muted-foreground">{label}</div>
    </div>
  )
}
