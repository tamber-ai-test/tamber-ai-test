'use client'

import { useState } from 'react'

export default function Home() {
  const [file, setFile] = useState<File | null>(null)

  const handleUpload = async () => {
    if (!file) return

    const formData = new FormData()
    formData.append('file', file)

    const res = await fetch('http://localhost:8080/upload', {
      method: 'POST',
      body: formData,
    })

    if (res.ok) alert('Uploaded!')
    else alert('Error uploading')
  }

  return (
    <div className="p-8">
      <input type="file" accept=".mp3" onChange={e => setFile(e.target.files?.[0] || null)} />
      <button className="mt-4 p-2 bg-blue-600 text-white" onClick={handleUpload}>Upload</button>
    </div>
  )
}

