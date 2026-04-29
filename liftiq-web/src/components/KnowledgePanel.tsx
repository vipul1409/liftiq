import { useState } from 'react';
import { queryKnowledge, type KnowledgeResponse } from '../api/knowledge';

export function KnowledgePanel() {
  const [question, setQuestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<KnowledgeResponse | null>(null);
  const [error, setError] = useState('');
  const [expanded, setExpanded] = useState(false);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!question.trim()) return;

    setLoading(true);
    setError('');
    setResult(null);

    try {
      const data = await queryKnowledge(question.trim(), 'Smartrise', 'C4');
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Knowledge base query failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="max-w-3xl mx-auto px-4 mb-4">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-4 py-2.5 bg-amber-50 border border-amber-200 rounded-lg hover:bg-amber-100 transition-colors"
      >
        <span className="text-sm font-semibold text-amber-800">
          Knowledge Base — Ask about fault codes, maintenance, wiring
        </span>
        <span className="text-amber-600 text-xs">{expanded ? 'Hide' : 'Show'}</span>
      </button>

      {expanded && (
        <div className="mt-2 p-4 bg-white border border-gray-200 rounded-lg shadow-sm">
          <form onSubmit={handleSearch} className="flex gap-2">
            <input
              type="text"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="e.g., What does fault code E-05 mean?"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              disabled={loading}
            />
            <button
              type="submit"
              disabled={loading || !question.trim()}
              className="px-4 py-2 bg-amber-600 text-white text-sm font-semibold rounded-lg hover:bg-amber-700 disabled:bg-amber-300 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? 'Asking...' : 'Ask'}
            </button>
          </form>

          {error && (
            <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-700 text-sm">{error}</p>
            </div>
          )}

          {result && (
            <div className="mt-3 space-y-3">
              {/* LLM Answer */}
              {result.answer && (
                <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg">
                  <p className="text-xs font-semibold text-amber-700 uppercase tracking-wide mb-1">
                    Answer
                  </p>
                  <p className="text-sm text-gray-800 whitespace-pre-wrap">{result.answer}</p>
                </div>
              )}

              {/* Source chunks */}
              {result.results.length > 0 && (
                <div>
                  <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">
                    Sources ({result.results.length})
                  </p>
                  {result.results.map((r) => (
                    <div
                      key={r.id}
                      className="mb-2 p-3 bg-gray-50 border border-gray-200 rounded-lg"
                    >
                      <div className="flex justify-between items-start mb-1">
                        <span className="text-xs font-semibold text-blue-700">
                          {r.section}
                        </span>
                        <span className="text-[10px] text-gray-400 shrink-0 ml-2">
                          {(r.similarity * 100).toFixed(0)}% match · p.{r.page_number}
                        </span>
                      </div>
                      <p className="text-xs text-gray-600 line-clamp-3">{r.content}</p>
                      <p className="text-[10px] text-gray-400 mt-1">
                        {r.document_name} · {r.controller_make} {r.controller_model}
                      </p>
                    </div>
                  ))}
                </div>
              )}

              {result.results.length === 0 && !result.answer && (
                <p className="text-sm text-gray-500 italic">
                  No relevant information found in the knowledge base.
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
