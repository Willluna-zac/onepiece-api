import type { Character } from '../api/client'
import { Link } from 'react-router-dom'

const HAKI_COLOR: Record<string, string> = {
  Conqueror:   'bg-purple-900 text-purple-300',
  Armament:    'bg-gray-800 text-gray-300',
  Observation: 'bg-blue-900 text-blue-300',
}

export function CharacterCard({ char }: { char: Character }) {
  return (
    <Link to={`/characters/${char.id}`} className="card block hover:scale-[1.02] transition-transform">
      {/* Imagen */}
      {char.imageUrl && (
        <div className="mb-3 flex justify-center">
          <img
            src={char.imageUrl}
            alt={char.name}
            className="h-36 w-auto object-contain rounded-lg"
            onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
          />
        </div>
      )}

      <div className="flex items-start justify-between mb-1">
        <h3 className="font-pirate text-gold text-xl leading-tight">{char.name}</h3>
      </div>

      <p className="text-straw/50 text-xs mb-1 italic">"{char.alias}"</p>
      <p className="text-straw/60 text-sm mb-3">
        {char.role} · {char.species}
      </p>

      <div className="flex flex-wrap gap-1.5 text-xs">
        {char.devilFruit && (
          <span className="badge bg-purple-900 text-purple-300">
            🍎 {char.devilFruit.name}
          </span>
        )}
        {(char.hakiAbilities ?? []).map(h => (
          <span key={h.hakiType} className={`badge ${HAKI_COLOR[h.hakiType] ?? 'bg-gray-700 text-gray-300'}`}>
            ⚡ {h.hakiType}{h.awakened ? ' ★' : ''}
          </span>
        ))}
      </div>
    </Link>
  )
}
