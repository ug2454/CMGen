import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  Typography,
  CircularProgress,
  Alert,
  Paper,
  Link,
} from '@mui/material';

interface Chapter {
  timestamp: number;
  title: string;
}

interface YouTubeExportProps {
  chapters: Chapter[];
}

export default function YouTubeExport({ chapters }: YouTubeExportProps) {
  const [videoId, setVideoId] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [helpDialogOpen, setHelpDialogOpen] = useState(false);

  const handleOpen = () => {
    setDialogOpen(true);
    setError(null);
    setSuccess(false);
  };

  const handleClose = () => {
    if (!loading) {
      setDialogOpen(false);
    }
  };

  const handleHelpOpen = () => {
    setHelpDialogOpen(true);
  };

  const handleHelpClose = () => {
    setHelpDialogOpen(false);
  };

  const extractVideoId = (url: string): string => {
    // Handle YouTube video URL formats
    // https://www.youtube.com/watch?v=VIDEO_ID
    // https://youtu.be/VIDEO_ID
    // https://www.youtube.com/embed/VIDEO_ID
    
    let videoId = '';
    
    try {
      if (url.includes('youtube.com/watch')) {
        const urlObj = new URL(url);
        videoId = urlObj.searchParams.get('v') || '';
      } else if (url.includes('youtu.be/')) {
        videoId = url.split('youtu.be/')[1].split('?')[0];
      } else if (url.includes('youtube.com/embed/')) {
        videoId = url.split('youtube.com/embed/')[1].split('?')[0];
      } else {
        // Assume the input is directly a video ID
        videoId = url;
      }
    } catch (err) {
      // If URL parsing fails, assume it's a direct video ID
      videoId = url;
    }
    
    return videoId;
  };

  const handleExport = async () => {
    if (!videoId.trim()) {
      setError('Please enter a valid YouTube video ID or URL');
      return;
    }

    const extractedVideoId = extractVideoId(videoId.trim());
    if (!extractedVideoId) {
      setError('Could not extract a valid video ID');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const response = await fetch('http://localhost:8080/api/youtube', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          videoId: extractedVideoId,
          chapters: chapters,
        }),
        // Set a longer timeout for the request
        signal: AbortSignal.timeout(60000), // 60 second timeout
      });

      if (!response.ok) {
        let errorMessage = 'Failed to update YouTube video';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorMessage;
        } catch (e) {
          // Use default error message if parsing fails
        }
        
        // Check for specific error cases
        if (errorMessage.includes('redirect_uri_mismatch')) {
          errorMessage = 'Authentication error: redirect URI mismatch. Please check the YouTube credentials setup in the backend.';
        } else if (errorMessage.includes('invalid_grant') || errorMessage.includes('authorization')) {
          errorMessage = 'Authentication error: The authorization token is invalid or expired. Please restart the backend and try again.';
        }
        
        throw new Error(errorMessage);
      }

      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const formatChaptersPreview = (): string => {
    return chapters.map((chapter, index) => {
      const minutes = Math.floor(chapter.timestamp / 60);
      const seconds = Math.floor(chapter.timestamp % 60);
      return `${minutes}:${seconds.toString().padStart(2, '0')} ${chapter.title}`;
    }).join('\n');
  };

  return (
    <Card sx={{ mt: 2 }}>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">YouTube Integration</Typography>
          <Box>
            <Button 
              variant="outlined"
              onClick={handleHelpOpen}
              sx={{ mr: 1 }}
            >
              Help
            </Button>
            <Button 
              variant="contained" 
              color="secondary" 
              onClick={handleOpen}
              disabled={chapters.length === 0}
            >
              Export to YouTube
            </Button>
          </Box>
        </Box>
        <Typography variant="body2" color="text.secondary">
          Export your chapters directly to a YouTube video description.
        </Typography>
      </CardContent>

      {/* Export Dialog */}
      <Dialog open={dialogOpen} onClose={handleClose} maxWidth="md" fullWidth>
        <DialogTitle>Export Chapters to YouTube</DialogTitle>
        <DialogContent>
          <DialogContentText gutterBottom>
            Enter your YouTube video ID or URL to update its description with these chapters.
            You must own the video or have edit access to it.
          </DialogContentText>
          
          <TextField
            autoFocus
            margin="dense"
            label="YouTube Video ID or URL"
            fullWidth
            variant="outlined"
            value={videoId}
            onChange={(e) => setVideoId(e.target.value)}
            disabled={loading}
            placeholder="https://www.youtube.com/watch?v=..."
            sx={{ mb: 2 }}
          />
          
          <Typography variant="subtitle2" gutterBottom>
            Preview of chapters to be added:
          </Typography>
          
          <Box 
            component="pre" 
            sx={{ 
              bgcolor: 'background.paper', 
              p: 2, 
              borderRadius: 1,
              maxHeight: '200px',
              overflow: 'auto',
              fontSize: '0.875rem',
              border: '1px solid',
              borderColor: 'divider'
            }}
          >
            {formatChaptersPreview()}
          </Box>
          
          {error && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {error}
              {(error.includes('redirect_uri_mismatch') || error.includes('authentication')) && (
                <Box sx={{ mt: 1 }}>
                  <Typography variant="body2">
                    Please check the README file for setup instructions or click the "Help" button.
                  </Typography>
                </Box>
              )}
            </Alert>
          )}
          
          {success && (
            <Alert severity="success" sx={{ mt: 2 }}>
              Successfully updated YouTube video description!
            </Alert>
          )}
          
          {loading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
              <CircularProgress />
              <Typography variant="body2" sx={{ ml: 2 }}>
                Updating YouTube video... This might take a moment.
              </Typography>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button 
            onClick={handleExport} 
            variant="contained" 
            color="primary"
            disabled={loading || !videoId.trim()}
          >
            {loading ? <CircularProgress size={24} /> : 'Export'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Help Dialog */}
      <Dialog open={helpDialogOpen} onClose={handleHelpClose} maxWidth="md">
        <DialogTitle>YouTube Integration Help</DialogTitle>
        <DialogContent>
          <Typography variant="h6" gutterBottom>Common Issues</Typography>
          
          <Paper elevation={0} variant="outlined" sx={{ p: 2, mb: 2 }}>
            <Typography variant="subtitle1" gutterBottom>Authentication Errors</Typography>
            <Typography variant="body2" paragraph>
              If you see errors like "redirect_uri_mismatch" or "invalid_grant", there's an issue with your YouTube API credentials.
            </Typography>
            <Typography variant="body2" component="div">
              To fix this:
              <ol>
                <li>Go to the <Link href="https://console.cloud.google.com/apis/credentials" target="_blank" rel="noopener">Google Cloud Console</Link></li>
                <li>Create new OAuth credentials of type <strong>Desktop application</strong> (not Web application)</li>
                <li>Download the credentials file and save it as <code>credentials.json</code> in your project root</li>
                <li>Restart the backend server</li>
              </ol>
            </Typography>
          </Paper>

          <Paper elevation={0} variant="outlined" sx={{ p: 2, mb: 2 }}>
            <Typography variant="subtitle1" gutterBottom>Long Loading Times</Typography>
            <Typography variant="body2">
              The first time you export to YouTube, the application needs to authenticate. A browser window should open to complete the authentication. If it doesn't open automatically, check your terminal for a link to open manually.
            </Typography>
          </Paper>

          <Paper elevation={0} variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" gutterBottom>Permission Issues</Typography>
            <Typography variant="body2">
              You can only update videos that you own or have editing rights to. Make sure you're logged into the correct YouTube/Google account.
            </Typography>
          </Paper>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleHelpClose}>Close</Button>
        </DialogActions>
      </Dialog>
    </Card>
  );
} 