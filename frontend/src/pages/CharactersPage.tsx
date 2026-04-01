import { useState } from 'react'
import { useCharacters, useSearch } from '../hooks/useApi'
import { CharacterCard } from '../components/CharacterCard'
import { Loader, ErrorMsg } from '../components/Loader'

export default function CharactersPage() {
  const [query, setQuery] = useState('')
  const { data: all, isLoading: loadingAll, error: errAll } = useCharacters()
  const { data: results, isLoading: loadingSearch } = useSearch(query)

  const isSearching = query.length > 1
  const characters = isSearching ? results : all
  const isLoading = isSearching ? loadingSearch : loadingAll

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="font-pirate text-gold text-4xl mb-2">Tripulación & Piratas</h1>
      <p className="text-straw/60 mb-6">Explora los personajes del mundo de One Piece</p>

      {/* Search */}
      <div className="relative mb-8">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-straw/40">🔍</span>
        <input
          type="text"
          placeholder="Buscar personaje…"
          value={query}
          onChange={e => setQuery(e.target.value)}
          className="w-full bg-navy-light border border-gold/20 focus:border-gold/60 rounded-xl pl-9 pr-4 py-3 text-straw placeholder-straw/30 outline-none transition-colors"
        />
      </div>

      {isLoading && <Loader />}
      {errAll && !isSearching && <ErrorMsg message={String(errAll)} />}

      {!isLoading && (
        <>
          <p className="text-straw/50 text-sm mb-4">
            {characters?.length ?? 0} personaje{characters?.length !== 1 ? 's' : ''}
          </p>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {characters?.map(c => <CharacterCard key={c.id} char={c} />)}
          </div>
          {isSearching && (!results || results.length === 0) && (
            <p className="text-center text-straw/40 py-16">No se encontró "{query}"</p>
          )}
        </>
      )}
    </div>
  )
}
