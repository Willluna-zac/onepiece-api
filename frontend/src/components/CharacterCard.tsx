import type { Character } from '../api/client'
import { Link } from 'react-router-dom'

const HAKI_COLOR: Record<string, string> = {
  Conqueror:   'bg-purple-900 text-purple-300',
  Armament:    'bg-gray-800 text-gray-300',
  Observation: 'bg-blue-900 text-blue-300',
}

export const FRUIT_EMOJI: Record<string, string> = {
  'Logia':         '🌋',
  'Mythical Zoan': '🐲',
  'Ancient Zoan':  '🦴',
  'Zoan':          '🦊',
  'Paramecia':     '🍇',
}

export function CharacterCard({ char }: { char: Character }) {
  const fruitEmoji = char.devilFruit ? (FRUIT_EMOJI[char.devilFruit.type] ?? '🍎') : null

  return (
    <Link to={`/characters/${char.id}`} className="card flex flex-col hover:scale-[1.02] transition-transform">
      {/* Imagen con fallback a avatar de iniciales */}
      <div className="mb-3 flex justify-center bg-navy-dark rounded-lg py-2 min-h-[140px] items-center">
        {char.imageUrl ? (
          <img
            src={char.imageUrl}
            alt={char.name}
            referrerPolicy="no-referrer"
            crossOrigin="anonymous"
            className="h-36 w-auto object-contain rounded-lg"
            onError={e => {
              const img = e.target as HTMLImageElement
              // Fallback: avatar con iniciales
              img.src = `https://ui-avatars.com/api/?name=${encodeURIComponent(char.name)}&background=1a2e45&color=f5c842&bold=true&size=128&font-size=0.33`
              img.referrerPolicy = 'origin'
            }}
          />
        ) : (
          <img
            src={`https://ui-avatars.com/api/?name=${encodeURIComponent(char.name)}&background=1a2e45&color=f5c842&bold=true&size=128&font-size=0.33`}
            alt={char.name}
            className="h-32 w-32 rounded-full"
          />
        )}
      </div>

      <div className="flex items-start justify-between mb-1">
        <h3 className="font-pirate text-gold text-xl leading-tight">{char.name}</h3>
      </div>

      <p className="text-straw/50 text-xs mb-1 italic truncate">"{char.alias}"</p>
      <p className="text-straw/60 text-sm mb-3">{char.role}</p>

      <div className="flex flex-wrap gap-1.5 text-xs mt-auto">
        {fruitEmoji && char.devilFruit && (
          <span className="badge bg-purple-900 text-purple-300">
            {fruitEmoji} {char.devilFruit.name}
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
