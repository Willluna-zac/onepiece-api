import { useParams, Link } from 'react-router-dom'
import { useCharacter } from '../hooks/useApi'
import { Loader, ErrorMsg } from '../components/Loader'
import { FRUIT_EMOJI } from '../components/CharacterCard'
import { proxyImage } from '../api/client'

const HAKI_COLOR: Record<string, string> = {
  Conqueror:   'border-purple-500 bg-purple-900/30',
  Armament:    'border-gray-500 bg-gray-800/30',
  Observation: 'border-blue-500 bg-blue-900/30',
}
const HAKI_ICON: Record<string, string> = {
  Conqueror: '👑', Armament: '🛡️', Observation: '👁️',
}

export default function CharacterDetailPage() {
  const { id = '' } = useParams()
  const { data: char, isLoading, error } = useCharacter(id)

  if (isLoading) return <Loader />
  if (error || !char) return <ErrorMsg message={error ? String(error) : 'Personaje no encontrado'} />

  const fruitEmoji = char.devilFruit ? (FRUIT_EMOJI[char.devilFruit.type] ?? '🍎') : null
  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(char.name)}&background=1a2e45&color=f5c842&bold=true&size=192&font-size=0.33`
  const imgSrc = proxyImage(char.imageUrl) ?? avatarUrl

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <Link to="/" className="text-gold hover:text-gold-light text-sm mb-6 inline-block">← Volver</Link>

      <div className="card mb-6">
        <div className="flex gap-6 flex-col sm:flex-row">
          <div className="flex justify-center">
            <img
              src={imgSrc}
              alt={char.name}
              className="h-48 w-auto object-contain rounded-xl"
              onError={e => { e.currentTarget.onerror = null; e.currentTarget.src = avatarUrl }}
            />
          </div>
          <div className="flex-1">
            <h1 className="font-pirate text-gold text-4xl mb-1">{char.name}</h1>
            <p className="text-straw/50 italic mb-4">"{char.alias}"</p>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div><p className="text-straw/40 text-xs">Rol</p><p className="font-semibold">{char.role}</p></div>
              <div><p className="text-straw/40 text-xs">Especie</p><p className="font-semibold">{char.species}</p></div>
              <div className="col-span-2"><p className="text-straw/40 text-xs">Primera aparición</p><p className="font-semibold">{char.firstAppearance}</p></div>
            </div>
            {char.notes && (
              <p className="text-straw/60 text-sm mt-4 border-l-2 border-gold/40 pl-3 italic">{char.notes}</p>
            )}
          </div>
        </div>
      </div>

      {char.devilFruit && (
        <div className="card mb-4">
          <h2 className="font-pirate text-gold text-2xl mb-3">{fruitEmoji} Fruta del Diablo</h2>
          <p className="font-bold text-lg">{char.devilFruit.name}</p>
          <span className="badge bg-purple-900 text-purple-300 mb-2 inline-block">{char.devilFruit.type}</span>
          <p className="text-straw/70 text-sm">{char.devilFruit.description}</p>
        </div>
      )}

      {(char.hakiAbilities ?? []).length > 0 && (
        <div className="card mb-4">
          <h2 className="font-pirate text-gold text-2xl mb-3">⚡ Haki</h2>
          <div className="flex flex-col gap-2">
            {char.hakiAbilities!.map(h => (
              <div key={h.hakiType} className={`rounded-lg px-4 py-3 border-l-4 ${HAKI_COLOR[h.hakiType] ?? 'border-gray-500'}`}>
                <div className="flex items-center gap-2 flex-wrap">
                  <span>{HAKI_ICON[h.hakiType]}</span>
                  <p className="font-semibold">{h.hakiType} Haki</p>
                  <span className="badge bg-navy text-straw/70 text-xs">{h.proficiency}</span>
                  {h.awakened && <span className="badge bg-gold text-navy text-xs">★ Despertado</span>}
                </div>
                {h.notes && <p className="text-straw/50 text-xs mt-1">{h.notes}</p>}
              </div>
            ))}
          </div>
        </div>
      )}

      {(char.abilities ?? []).length > 0 && (
        <div className="card">
          <h2 className="font-pirate text-gold text-2xl mb-3">⚔️ Habilidades</h2>
          <ul className="space-y-2">
            {char.abilities!.map(a => (
              <li key={a.type} className="border-l-2 border-gold/40 pl-3">
                <p className="font-semibold text-sm">{a.type}</p>
                {a.notes && <p className="text-straw/60 text-xs">{a.notes}</p>}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
