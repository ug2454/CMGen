import React, { useState } from 'react';
import { Box, Container, Snackbar, Alert } from '@mui/material';
import VideoProcessor from './components/VideoProcessor';
import ChapterEditor from './components/ChapterEditor';

interface Chapter {
  timestamp: number;
  title: string;
}

export default function App() {
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [message, setMessage] = useState<{ text: string; severity: 'success' | 'error' } | null>(null);

  const handleProcessingStart = () => {
    setMessage({ text: 'Processing video...', severity: 'success' });
  };

  const handleProcessingComplete = (newChapters: Chapter[]) => {
    setChapters(newChapters);
    setMessage({ text: 'Video processed successfully!', severity: 'success' });
  };

  const handleChaptersChange = (newChapters: Chapter[]) => {
    setChapters(newChapters);
  };

  const handleExport = async () => {
    try {
      const response = await fetch('/api/export');
      if (!response.ok) throw new Error('Export failed');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'chapters.json';
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      setMessage({ text: 'Chapters exported successfully!', severity: 'success' });
    } catch (err) {
      setMessage({ text: 'Failed to export chapters', severity: 'error' });
    }
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        <VideoProcessor
          onProcessingStart={handleProcessingStart}
          onProcessingComplete={handleProcessingComplete}
        />
        {chapters.length > 0 && (
          <ChapterEditor
            chapters={chapters}
            onChaptersChange={handleChaptersChange}
            onExport={handleExport}
          />
        )}
      </Box>

      <Snackbar
        open={!!message}
        autoHideDuration={6000}
        onClose={() => setMessage(null)}
      >
        <Alert
          onClose={() => setMessage(null)}
          severity={message?.severity}
          sx={{ width: '100%' }}
        >
          {message?.text}
        </Alert>
      </Snackbar>
    </Container>
  );
} 