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
  IconButton,
  List,
  ListItem,
  Paper,
  TextField,
  Typography,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

interface Chapter {
  timestamp: number;
  title: string;
}

interface ChapterEditorProps {
  chapters: Chapter[];
  onChaptersChange: (chapters: Chapter[]) => void;
  onExport: () => void;
}

export default function ChapterEditor({ chapters, onChaptersChange, onExport }: ChapterEditorProps) {
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [chapterToDelete, setChapterToDelete] = useState<number | null>(null);

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const moveChapter = (fromIndex: number, toIndex: number) => {
    const updatedChapters = [...chapters];
    const [movedItem] = updatedChapters.splice(fromIndex, 1);
    updatedChapters.splice(toIndex, 0, movedItem);
    onChaptersChange(updatedChapters);
  };

  const handleTitleChange = (index: number, newTitle: string) => {
    const newChapters = [...chapters];
    newChapters[index] = { ...newChapters[index], title: newTitle };
    onChaptersChange(newChapters);
  };

  const confirmDelete = (index: number) => {
    setChapterToDelete(index);
    setDeleteDialogOpen(true);
  };

  const handleDelete = () => {
    if (chapterToDelete !== null) {
      const newChapters = chapters.filter((_, i) => i !== chapterToDelete);
      onChaptersChange(newChapters);
      setDeleteDialogOpen(false);
      setChapterToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setChapterToDelete(null);
  };

  const startEditing = (index: number) => {
    setEditingIndex(index);
    setEditTitle(chapters[index].title);
  };

  const finishEditing = () => {
    if (editingIndex !== null) {
      handleTitleChange(editingIndex, editTitle);
      setEditingIndex(null);
    }
  };

  const moveUp = (index: number) => {
    if (index > 0) {
      moveChapter(index, index - 1);
    }
  };

  const moveDown = (index: number) => {
    if (index < chapters.length - 1) {
      moveChapter(index, index + 1);
    }
  };

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">Chapters</Typography>
          <Button variant="contained" onClick={onExport}>
            Export Chapters
          </Button>
        </Box>

        <Paper variant="outlined" sx={{ maxHeight: '500px', overflow: 'auto' }}>
          <List>
            {chapters.map((chapter, index) => (
              <ListItem
                key={`chapter-${index}`}
                divider
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  padding: 2,
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', mr: 1 }}>
                    <IconButton 
                      size="small" 
                      onClick={() => moveUp(index)}
                      disabled={index === 0}
                    >
                      ↑
                    </IconButton>
                    <IconButton 
                      size="small" 
                      onClick={() => moveDown(index)}
                      disabled={index === chapters.length - 1}
                    >
                      ↓
                    </IconButton>
                  </Box>
                  
                  <Box sx={{ width: '80px', mr: 2 }}>
                    <Typography variant="body2">
                      {formatTime(chapter.timestamp)}
                    </Typography>
                  </Box>
                  
                  <Box sx={{ flex: 1 }}>
                    {editingIndex === index ? (
                      <TextField
                        fullWidth
                        variant="outlined"
                        size="small"
                        value={editTitle}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditTitle(e.target.value)}
                        onBlur={finishEditing}
                        onKeyPress={(e: React.KeyboardEvent<HTMLInputElement>) => {
                          if (e.key === 'Enter') {
                            finishEditing();
                          }
                        }}
                        autoFocus
                      />
                    ) : (
                      <Typography
                        onClick={() => startEditing(index)}
                        sx={{ cursor: 'pointer' }}
                      >
                        {chapter.title}
                      </Typography>
                    )}
                  </Box>
                  
                  <Box>
                    <IconButton
                      edge="end"
                      onClick={() => confirmDelete(index)}
                      size="small"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                </Box>
              </ListItem>
            ))}
          </List>
        </Paper>

        {/* Delete Confirmation Dialog */}
        <Dialog
          open={deleteDialogOpen}
          onClose={handleCancelDelete}
          aria-labelledby="alert-dialog-title"
          aria-describedby="alert-dialog-description"
        >
          <DialogTitle id="alert-dialog-title">
            Confirm Deletion
          </DialogTitle>
          <DialogContent>
            <DialogContentText id="alert-dialog-description">
              Are you sure you want to delete this chapter?
              {chapterToDelete !== null && chapters[chapterToDelete] && (
                <>
                  <br />
                  <strong>Time:</strong> {formatTime(chapters[chapterToDelete].timestamp)}
                  <br />
                  <strong>Title:</strong> {chapters[chapterToDelete].title}
                </>
              )}
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCancelDelete} color="primary">
              Cancel
            </Button>
            <Button onClick={handleDelete} color="error" autoFocus>
              Delete
            </Button>
          </DialogActions>
        </Dialog>
      </CardContent>
    </Card>
  );
} 