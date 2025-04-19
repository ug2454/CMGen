/// <reference types="react-scripts" />

declare module 'react-beautiful-dnd' {
  export const Droppable: any;
  export const Draggable: any;
  export const DragDropContext: any;
  export interface DropResult {
    destination: {
      index: number;
    } | null;
    source: {
      index: number;
    };
  }
  export interface DroppableProvided {
    innerRef: React.RefObject<any>;
    droppableProps: any;
    placeholder: React.ReactNode;
  }
  export interface DraggableProvided {
    innerRef: React.RefObject<any>;
    draggableProps: any;
    dragHandleProps: any;
  }
}

declare module '@mui/material/styles' {
  export const styled: any;
  export const ThemeProvider: React.ComponentType<{
    theme: any;
    children: React.ReactNode;
  }>;
  export function createTheme(options?: any): any;
}

declare module '@mui/icons-material/Delete' {
  const DeleteIcon: React.ComponentType<any>;
  export default DeleteIcon;
}

declare module '@mui/icons-material/DragHandle' {
  const DragHandleIcon: React.ComponentType<any>;
  export default DragHandleIcon;
}

declare module '@mui/material/CssBaseline' {
  const CssBaseline: React.ComponentType<any>;
  export default CssBaseline;
} 