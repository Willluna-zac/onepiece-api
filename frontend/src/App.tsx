import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Navbar } from './components/Navbar'
import CharactersPage from './pages/CharactersPage'
import CharacterDetailPage from './pages/CharacterDetailPage'
import DevilFruitsPage from './pages/DevilFruitsPage'
import IslandsPage from './pages/IslandsPage'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen">
          <Navbar />
          <main>
            <Routes>
              <Route path="/" element={<CharactersPage />} />
              <Route path="/characters/:id" element={<CharacterDetailPage />} />
              <Route path="/devil-fruits" element={<DevilFruitsPage />} />
              <Route path="/islands" element={<IslandsPage />} />
            </Routes>
          </main>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
