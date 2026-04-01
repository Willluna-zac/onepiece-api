export function Loader() {
  return (
    <div className="flex items-center justify-center py-20">
      <div className="w-10 h-10 border-4 border-gold/30 border-t-gold rounded-full animate-spin" />
    </div>
  )
}

export function ErrorMsg({ message }: { message: string }) {
  return (
    <div className="text-center py-16 text-red-400">
      <p className="text-4xl mb-3">☠️</p>
      <p className="font-semibold">{message}</p>
    </div>
  )
}
