import { useState } from "react"

function TodoList() {
    
  const [todos, setTodos] = useState([
    { id: 1, text: '買牛奶' },
    { id: 2, text: '寫 React' },
    { id: 3, text: '運動' },
  ]);

  const handleDelete = (id) => {
    setTodos(prev => prev.filter(todo => todo.id !== id));
  };

  return (
    <ul>
      {todos.map(todo => (
        <li key={todo.id}>
          {todo.text}
          <button onClick={() => handleDelete(todo.id)}>刪除</button>
        </li>
      ))}
    </ul>
  );
}
export default TodoList;