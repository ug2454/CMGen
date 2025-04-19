import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';
import DeleteIcon from '@mui/icons-material/Delete';
import DragHandleIcon from '@mui/icons-material/DragHandle';

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

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const handleDragEnd = (result: any) => {
    if (!result.destination) return;

    const items = Array.from(chapters);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    onChaptersChange(items);
  };

  const handleTitleChange = (index: number, newTitle: string) => {
    const newChapters = [...chapters];
    newChapters[index] = { ...newChapters[index], title: newTitle };
    onChaptersChange(newChapters);
  };

  const handleDelete = (index: number) => {
    const newChapters = chapters.filter((_, i) => i !== index);
    onChaptersChange(newChapters);
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

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">Chapters</Typography>
          <Button variant="contained" onClick={onExport}>
            Export Chapters
          </Button>
        </Box>

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell width={50}></TableCell>
                <TableCell width={100}>Time</TableCell>
                <TableCell>Title</TableCell>
                <TableCell width={100} align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <DragDropContext onDragEnd={handleDragEnd}>
              <Droppable droppableId="chapters">
                {(provided) => (
                  <TableBody
                    {...provided.droppableProps}
                    ref={provided.innerRef}
                  >
                    {chapters.map((chapter, index) => (
                      <Draggable
                        key={index}
                        draggableId={`chapter-${index}`}
                        index={index}
                      >
                        {(provided) => (
                          <TableRow
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                          >
                            <TableCell {...provided.dragHandleProps}>
                              <DragHandleIcon />
                            </TableCell>
                            <TableCell>{formatTime(chapter.timestamp)}</TableCell>
                            <TableCell>
                              {editingIndex === index ? (
                                <TextField
                                  fullWidth
                                  value={editTitle}
                                  onChange={(e) => setEditTitle(e.target.value)}
                                  onBlur={finishEditing}
                                  onKeyPress={(e) => {
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
                            </TableCell>
                            <TableCell align="right">
                              <IconButton
                                onClick={() => handleDelete(index)}
                                size="small"
                              >
                                <DeleteIcon />
                              </IconButton>
                            </TableCell>
                          </TableRow>
                        )}
                      </Draggable>
                    ))}
                    {provided.placeholder}
                  </TableBody>
                )}
              </Droppable>
            </DragDropContext>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
} 