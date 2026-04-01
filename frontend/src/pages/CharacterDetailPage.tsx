import { useParams, Link } from 'react-router-dom'
import { useCharacter } from '../hooks/useApi'
import { Loader, ErrorMsg } from '../components/Loader'

function bountyLabel(bounty: number) {
  if (bounty >= 1_000_000_000) return `${(bounty / 1_000_000_000).toFixed(1)}B`
  if (bounty >= 1_000_000) return `${(bounty / 1_000_000).toFixed(0)}M`
  return bounty.toLocaleString()
}

export default function CharacterDetailPage() {
  const { id = '' } = useParams()
  const { data: char, isLoading, error } = useCharacter(id)

  if (isLoading) return <Loader />
  if (error || !char) return <ErrorMsg message={error ? String(error) : 'Personaje no encontrado'} />

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <Link to="/" className="text-gold hover:text-gold-light text-sm mb-6 inline-block">← Volver</Link>

      <div className="card mb-6">
        <div className="flex items-start justify-between mb-1">
          <h1 className="font-pirate text-gold text-4xl">{char.name}</h1>
          <span className={`badge text-sm ${char.status === 'Alive' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'}`}>
            {char.status}
          </span>
        </div>
        <p className="text-straw/60 mb-4">{char.role} · {char.crew}</p>

        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-sm">
          <div><span className="text-straw/40">Origen</span><p className="font-semibold">{char.origin}</p></div>
          <div><span className="text-straw/40">Edad</span><p className="font-semibold">{char.age} años</p></div>
          {char.bounty > 0 && (
            <div><span className="text-straw/40">Recompensa</span><p className="font-semibold text-gold">💰 {bountyLabel(char.bounty)} Berry</p></div>
          )}
        </div>
      </div>

      {char.devilFruit && (
        <div className="card mb-4">
          <h2 className="font-pirate text-gold text-2xl mb-3">🍎 Fruta del Diablo</h2>
          <p className="font-bold text-lg">{char.devilFruit.name}</p>
          <span className="badge bg-purple-900 text-purple-300 mb-2 inline-block">{char.devilFruit.type}</span>
          <p className="text-straw/70 text-sm">{char.devilFruit.description}</p>
        </div>
      )}

      {(char.hakiAbilities ?? []).length > 0 && (
        <div className="card mb-4">
          <h2 className="font-pirate text-gold text-2xl mb-3">⚡ Haki</h2>
          <div className="flex flex-wrap gap-2">
            {char.hakiAbilities!.map(h => (
              <div key={h.type} className="bg-gray-800 rounded-lg px-3 py-2">
                <p className="font-semibold text-sm">{h.type}</p>
                <p className="text-straw/50 text-xs">{h.level}</p>
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
              <li key={a.name} className="border-l-2 border-gold/40 pl-3">
                <p className="font-semibold text-sm">{a.name}</p>
                <p className="text-straw/60 text-xs">{a.description}</p>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
