import type { Character } from '../api/client'
import { Link } from 'react-router-dom'

function bountyLabel(bounty: number) {
  if (bounty >= 1_000_000_000) return `${(bounty / 1_000_000_000).toFixed(1)}B`
  if (bounty >= 1_000_000) return `${(bounty / 1_000_000).toFixed(0)}M`
  return bounty.toLocaleString()
}

export function CharacterCard({ char }: { char: Character }) {
  return (
    <Link to={`/characters/${char.id}`} className="card block hover:scale-[1.02] transition-transform">
      <div className="flex items-start justify-between mb-2">
        <h3 className="font-pirate text-gold text-xl leading-tight">{char.name}</h3>
        <span className={`badge ${char.status === 'Alive' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'}`}>
          {char.status}
        </span>
      </div>

      <p className="text-straw/60 text-sm mb-3">
        {char.role} · {char.crew}
      </p>

      <div className="flex flex-wrap gap-2 text-xs">
        {char.devilFruit && (
          <span className="badge bg-purple-900 text-purple-300">
            🍎 {char.devilFruit.name}
          </span>
        )}
        {char.bounty > 0 && (
          <span className="badge bg-gold/20 text-gold">
            💰 {bountyLabel(char.bounty)} Berry
          </span>
        )}
        {(char.hakiAbilities ?? []).map(h => (
          <span key={h.type} className="badge bg-gray-800 text-gray-300">
            ⚡ {h.type}
          </span>
        ))}
      </div>
    </Link>
  )
}
