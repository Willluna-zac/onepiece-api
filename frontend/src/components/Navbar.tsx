import { Link, useLocation } from 'react-router-dom'

const links = [
  { to: '/', label: '🏴‍☠️ Personajes' },
  { to: '/devil-fruits', label: '🍎 Devil Fruits' },
  { to: '/islands', label: '🗺️ Mapa de Islas' },
  { to: '/routes', label: '⚓ Rutas' },
  { to: '/stats', label: '📊 Análisis' },
]

export function Navbar() {
  const { pathname } = useLocation()
  return (
    <nav className="sticky top-0 z-50 bg-navy-dark border-b border-gold/30 shadow-lg">
      <div className="max-w-6xl mx-auto px-4 py-3 flex items-center gap-8">
        <Link to="/" className="font-pirate text-gold text-2xl tracking-wide hover:text-gold-light transition-colors">
          ☠ ONE PIECE API
        </Link>
        <div className="flex gap-2 ml-auto">
          {links.map(({ to, label }) => (
            <Link
              key={to}
              to={to}
              className={`px-3 py-1.5 rounded-lg text-sm font-semibold transition-colors ${
                pathname === to
                  ? 'bg-gold text-navy'
                  : 'text-straw hover:bg-navy-light'
              }`}
            >
              {label}
            </Link>
          ))}
        </div>
      </div>
    </nav>
  )
}
