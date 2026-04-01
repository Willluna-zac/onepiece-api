import type { Character } from '../api/client'
import { proxyImage } from '../api/client'
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
  const imgSrc = proxyImage(char.imageUrl)
  const avatarSrc = `https://ui-avatars.com/api/?name=${encodeURIComponent(char.name)}&background=1a2e45&color=f5c842&bold=true&size=128&font-size=0.33`

  return (
    <Link to={`/characters/${char.id}`} className="card flex flex-col hover:scale-[1.02] transition-transform">
      <div className="mb-3 flex justify-center bg-navy-dark rounded-lg py-2 min-h-[140px] items-center">
        <img
          src={imgSrc ?? avatarSrc}
          alt={char.name}
          className="h-36 w-auto max-w-full object-contain rounded-lg"
          onError={e => {
            const el = e.currentTarget
            el.onerror = null
            el.src = avatarSrc
          }}
        />
      </div>

      <h3 className="font-pirate text-gold text-xl leading-tight mb-0.5">{char.name}</h3>
      <p className="text-straw/50 text-xs italic truncate mb-1">"{char.alias}"</p>
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
