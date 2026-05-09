'use client'

import Link from 'next/link'
import { Button } from '@/components/ui/button'

export default function RegisterForm() {
  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card p-4 text-sm text-muted-foreground">
        Self-registration is disabled. Ask an administrator to create your account and assign your 50-image quota.
      </div>
      <Button asChild className="w-full">
        <Link href="/login">Back to Login</Link>
      </Button>
    </div>
  )
}
