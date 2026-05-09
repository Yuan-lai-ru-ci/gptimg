import LoginForm from '@/components/auth/LoginForm'

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-foreground">GPT Image</h1>
          <p className="text-muted-foreground mt-2">Sign in to start generating images</p>
        </div>
        <div className="bg-card border border-border rounded-2xl p-6">
          <LoginForm />
        </div>
      </div>
    </div>
  )
}
