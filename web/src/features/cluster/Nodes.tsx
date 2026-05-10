export default function Nodes() {
  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white mb-2">Node Explorer</h1>
        <p className="text-neutral-400">Detailed health and metrics for all physical compute nodes.</p>
      </div>
      <div className="flex flex-col items-center justify-center p-24 border border-dashed border-neutral-800 rounded-xl bg-neutral-900/20">
        <h3 className="text-lg font-medium text-white">Under Construction</h3>
        <p className="text-neutral-400 mt-2 max-w-md text-center">The detailed node explorer is currently being built. Check back soon for node-level GPU heatmaps and lifecycle management.</p>
      </div>
    </div>
  )
}
