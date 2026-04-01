import { useDevilFruitChars } from '../hooks/useApi'
import { CharacterCard } from '../components/CharacterCard'
import { Loader, ErrorMsg } from '../components/Loader'

export default function DevilFruitsPage() {
  const { data, isLoading, error } = useDevilFruitChars()

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="font-pirate text-gold text-4xl mb-2">🍎 Frutas del Diablo</h1>
      <p className="text-straw/60 mb-6">Personajes que poseen una Akuma no Mi</p>

      {isLoading && <Loader />}
      {error && <ErrorMsg message={String(error)} />}

      {!isLoading && data && (
        <>
          <p className="text-straw/50 text-sm mb-4">{data.length} usuarios encontrados</p>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {data.map(c => <CharacterCard key={c.id} char={c} />)}
          </div>
        </>
      )}
    </div>
  )
}
