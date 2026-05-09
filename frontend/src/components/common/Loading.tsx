interface LoadingProps {
  label?: string
  className?: string
}

export default function Loading({ label, className = '' }: LoadingProps) {
  return (
    <div className={`flex flex-col items-center justify-center gap-3 ${className}`.trim()}>
      <div className="relative h-10 w-10">
        <div className="absolute inset-0 animate-ping rounded-full border border-sky-400/40"></div>
        <div className="absolute inset-1 rounded-full border border-cyan-300/70"></div>
        <div className="absolute inset-[10px] animate-spin rounded-full border-2 border-transparent border-t-white border-r-sky-400"></div>
      </div>
      {label ? <p className="text-sm text-muted-foreground">{label}</p> : null}
    </div>
  )
}
