import { render, screen } from '@testing-library/react';
import App from './App';

beforeEach(() => {
  localStorage.clear();
});

describe('App routing', () => {
  test('renders login page when no token', () => {
    render(<App />);
    const heading = screen.getByRole('heading', { name: /login/i });
    expect(heading).toBeInTheDocument();
  });

  test('redirects to home page when token exists', () => {
    localStorage.setItem('accessToken', 'test-token');
    render(<App />);
    const homeText = screen.getByText(/welcome to the home page/i);
    expect(homeText).toBeInTheDocument();
  });
});
