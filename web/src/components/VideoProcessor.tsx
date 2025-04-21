import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  FormControl,
  FormHelperText,
  InputLabel,
  OutlinedInput,
  Slider,
  Typography,
} from '@mui/material';
import { styled } from '@mui/material/styles';

const Input = styled('input')({
  display: 'none',
});

interface VideoProcessorProps {
  onProcessingStart: () => void;
  onProcessingComplete: (chapters: any[]) => void;
}

export default function VideoProcessor({ onProcessingStart, onProcessingComplete }: VideoProcessorProps) {
  const [file, setFile] = useState<File | null>(null);
  const [threshold, setThreshold] = useState(0.3);
  const [minGap, setMinGap] = useState(5.0);
  const [minDuration, setMinDuration] = useState(0.0);
  const [maxScenes, setMaxScenes] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setFile(event.target.files[0]);
      setError(null);
    }
  };

  const handleProcess = async () => {
    if (!file) {
      setError('Please select a video file');
      return;
    }

    setIsProcessing(true);
    onProcessingStart();
    setError(null);

    const formData = new FormData();
    formData.append('video', file);
    formData.append('threshold', threshold.toString());
    formData.append('minGap', minGap.toString());
    formData.append('minDuration', minDuration.toString());
    formData.append('maxScenes', maxScenes.toString());

    try {
      const response = await fetch('http://localhost:8080/api/detect', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to process video');
      }

      const chapters = await response.json();
      onProcessingComplete(chapters);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Video Processing
        </Typography>

        <Box sx={{ mb: 3 }}>
          <label htmlFor="video-upload">
            <Input
              accept="video/*"
              id="video-upload"
              type="file"
              onChange={handleFileChange}
            />
            <Button
              variant="contained"
              component="span"
              disabled={isProcessing}
            >
              Select Video
            </Button>
          </label>
          {file && (
            <Typography variant="body2" sx={{ mt: 1 }}>
              Selected: {file.name}
            </Typography>
          )}
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography gutterBottom>Scene Detection Threshold</Typography>
          <Slider
            value={threshold}
            onChange={(_, value) => setThreshold(Array.isArray(value) ? value[0] : value)}
            min={0}
            max={1}
            step={0.1}
            marks
            valueLabelDisplay="auto"
            disabled={isProcessing}
          />
          <FormHelperText>
            Higher values detect fewer but more significant scene changes
          </FormHelperText>
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography gutterBottom>Minimum Gap Between Scenes (seconds)</Typography>
          <Slider
            value={minGap}
            onChange={(_, value) => setMinGap(Array.isArray(value) ? value[0] : value)}
            min={0}
            max={30}
            step={1}
            marks
            valueLabelDisplay="auto"
            disabled={isProcessing}
          />
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography gutterBottom>Minimum Scene Duration (seconds)</Typography>
          <Slider
            value={minDuration}
            onChange={(_, value) => setMinDuration(Array.isArray(value) ? value[0] : value)}
            min={0}
            max={60}
            step={1}
            marks
            valueLabelDisplay="auto"
            disabled={isProcessing}
          />
        </Box>

        <Box sx={{ mb: 3 }}>
          <FormControl fullWidth>
            <InputLabel htmlFor="max-scenes">Maximum Number of Scenes</InputLabel>
            <OutlinedInput
              id="max-scenes"
              type="number"
              value={maxScenes}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setMaxScenes(parseInt(e.target.value) || 0)}
              disabled={isProcessing}
            />
            <FormHelperText>
              Set to 0 for unlimited scenes
            </FormHelperText>
          </FormControl>
        </Box>

        {error && (
          <Typography color="error" sx={{ mb: 2 }}>
            {error}
          </Typography>
        )}

        <Button
          variant="contained"
          color="primary"
          onClick={handleProcess}
          disabled={!file || isProcessing}
          fullWidth
        >
          {isProcessing ? 'Processing...' : 'Process Video'}
        </Button>
      </CardContent>
    </Card>
  );
} 