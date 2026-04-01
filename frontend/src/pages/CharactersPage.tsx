import { useState, useMemo } from 'react'
import { useCharacters } from '../hooks/useApi'
import { CharacterCard } from '../components/CharacterCard'
import { Loader, ErrorMsg } from '../components/Loader'

const PAGE_SIZE = 12

export default function CharactersPage() {
  const [query, setQuery]   = useState('')
  const [page, setPage]     = useState(1)
  const { data: all, isLoading, error } = useCharacters()

  // Búsqueda client-side: substring case-insensitive sobre nombre y alias
  const filtered = useMemo(() => {
    if (!all) return []
    const q = query.trim().toLowerCase()
    if (!q) return all
    return all.filter(c =>
      c.name.toLowerCase().includes(q) ||
      c.alias.toLowerCase().includes(q) ||
      c.role.toLowerCase().includes(q) ||
      (c.devilFruit?.name.toLowerCase().includes(q) ?? false)
    )
  }, [all, query])

  // Reset página al buscar
  const handleSearch = (v: string) => { setQuery(v); setPage(1) }

  // Paginación
  const totalPages = Math.ceil(filtered.length / PAGE_SIZE)
  const paginated  = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="font-pirate text-gold text-4xl mb-1">Tripulación & Piratas</h1>
      <p className="text-straw/60 mb-6">Explora los {all?.length ?? '…'} personajes del mundo de One Piece</p>

      {/* Search */}
      <div className="relative mb-8">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-straw/40 text-lg">🔍</span>
        <input
          type="text"
          placeholder="Busca por nombre, alias, rol o fruta del diablo…"
          value={query}
          onChange={e => handleSearch(e.target.value)}
          className="w-full bg-navy-light border border-gold/20 focus:border-gold/60 rounded-xl pl-10 pr-10 py-3 text-straw placeholder-straw/30 outline-none transition-colors"
        />
        {query && (
          <button
            onClick={() => handleSearch('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-straw/40 hover:text-straw text-lg leading-none"
          >✕</button>
        )}
      </div>

      {isLoading && <Loader />}
      {error && <ErrorMsg message={String(error)} />}

      {!isLoading && !error && (
        <>
          {/* Conteo */}
          <p className="text-straw/50 text-sm mb-4">
            {filtered.length} personaje{filtered.length !== 1 ? 's' : ''}
            {query && ` para "${query}"`}
            {totalPages > 1 && ` — página ${page} de ${totalPages}`}
          </p>

          {/* Grid */}
          {paginated.length > 0 ? (
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {paginated.map(c => <CharacterCard key={c.id} char={c} />)}
            </div>
          ) : (
            <div className="text-center py-20 text-straw/40">
              <p className="text-5xl mb-4">☠️</p>
              <p className="text-lg">No se encontró "{query}"</p>
              <p className="text-sm mt-1">Intenta con otro nombre o alias</p>
            </div>
          )}

          {/* Paginación */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-10">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn-primary disabled:opacity-30 disabled:cursor-not-allowed px-5"
              >← Anterior</button>

              <div className="flex gap-1">
                {Array.from({ length: totalPages }, (_, i) => i + 1).map(p => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={`w-9 h-9 rounded-lg text-sm font-bold transition-colors ${
                      p === page
                        ? 'bg-gold text-navy'
                        : 'bg-navy-light text-straw/60 hover:text-straw'
                    }`}
                  >{p}</button>
                ))}
              </div>

              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="btn-primary disabled:opacity-30 disabled:cursor-not-allowed px-5"
              >Siguiente →</button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
