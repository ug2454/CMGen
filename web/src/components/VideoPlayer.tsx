import React, { useRef, useState } from 'react';
import ReactPlayer from 'react-player';
import { Box, IconButton, Slider, Typography } from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';

interface VideoPlayerProps {
  url: string;
  currentTime: number;
  onTimeUpdate: (time: number) => void;
}

export const VideoPlayer: React.FC<VideoPlayerProps> = ({ url, currentTime, onTimeUpdate }) => {
  const playerRef = useRef<ReactPlayer>(null);
  const [playing, setPlaying] = useState(false);
  const [duration, setDuration] = useState(0);

  const handleProgress = (state: { playedSeconds: number }) => {
    onTimeUpdate(state.playedSeconds);
  };

  const handleDuration = (duration: number) => {
    setDuration(duration);
  };

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  return (
    <Box sx={{ width: '100%', maxWidth: 800, mx: 'auto', my: 2 }}>
      <Box sx={{ position: 'relative', paddingTop: '56.25%' }}>
        <ReactPlayer
          ref={playerRef}
          url={url}
          playing={playing}
          controls={true}
          width="100%"
          height="100%"
          style={{ position: 'absolute', top: 0, left: 0 }}
          onProgress={handleProgress}
          onDuration={handleDuration}
        />
      </Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
        <IconButton onClick={() => setPlaying(!playing)}>
          {playing ? <PauseIcon /> : <PlayArrowIcon />}
        </IconButton>
        <Typography sx={{ mx: 2 }}>{formatTime(currentTime)}</Typography>
        <Slider
          value={currentTime}
          max={duration}
          onChange={(_, value) => {
            if (typeof value === 'number') {
              onTimeUpdate(value);
              playerRef.current?.seekTo(value);
            }
          }}
          sx={{ flexGrow: 1 }}
        />
        <Typography sx={{ mx: 2 }}>{formatTime(duration)}</Typography>
      </Box>
    </Box>
  );
}; 