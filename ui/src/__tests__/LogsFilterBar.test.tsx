/**
 * LogsFilterBar Component Test
 */

import { describe, it, expect, jest } from '@jest/globals';
import { render, screen, fireEvent } from '@testing-library/react';
import { LogsFilterBar } from '../tabs/LogsFilterBar';

describe('LogsFilterBar', () => {
  const defaultProps = {
    namespace: 'default',
    containers: [
      { name: 'container-1' },
      { name: 'container-2' },
    ],
    selectedContainer: 'container-1',
    onContainerChange: jest.fn(),
    tailLines: '100',
    onTailLinesChange: jest.fn(),
    timestamps: false,
    onTimestampsChange: jest.fn(),
    previous: false,
    onPreviousChange: jest.fn(),
    wrapLines: true,
    onWrapLinesChange: jest.fn(),
    showSettings: false,
    onToggleSettings: jest.fn(),
    searchTerm: '',
    onSearchChange: jest.fn(),
    onSearch: jest.fn(),
    onClear: jest.fn(),
    onDownload: jest.fn(),
    connected: true,
    logCount: 150,
    totalLogs: 200,
  };

  it('should render LogsFilterBar', () => {
    render(<LogsFilterBar {...defaultProps} />);

    expect(screen.getByText('Logs')).toBeInTheDocument();
    expect(screen.getByText('150 / 200')).toBeInTheDocument();
  });

  it('should render container selector when multiple containers', () => {
    render(<LogsFilterBar {...defaultProps} />);

    expect(screen.getByText('container-1')).toBeInTheDocument();
  });

  it('should not render container selector when single container', () => {
    render(
      <LogsFilterBar
        {...defaultProps}
        containers={[{ name: 'single-container' }]}
      />
    );

    expect(screen.queryByText('single-container')).not.toBeInTheDocument();
  });

  it('should call onContainerChange when container selected', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const containerSelect = screen.getByText('container-1');
    fireEvent.click(containerSelect);

    expect(defaultProps.onContainerChange).not.toHaveBeenCalled();
  });

  it('should render search input', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search logs...');
    expect(searchInput).toBeInTheDocument();
  });

  it('should call onSearchChange when typing in search', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search logs...');
    fireEvent.change(searchInput, { target: { value: 'test' } });

    expect(defaultProps.onSearchChange).toHaveBeenCalledWith('test');
  });

  it('should call onSearch when pressing Enter', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search logs...');
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.keyDown(searchInput, { key: 'Enter' });

    expect(defaultProps.onSearch).toHaveBeenCalled();
  });

  it('should render settings button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const settingsBtn = screen.getByTitle('Settings');
    expect(settingsBtn).toBeInTheDocument();
  });

  it('should call onToggleSettings when clicking settings button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const settingsBtn = screen.getByTitle('Settings');
    fireEvent.click(settingsBtn);

    expect(defaultProps.onToggleSettings).toHaveBeenCalled();
  });

  it('should show settings panel when showSettings is true', () => {
    render(<LogsFilterBar {...defaultProps} showSettings={true} />);

    expect(screen.getByText('Initial Lines')).toBeInTheDocument();
    expect(screen.getByText('Show Timestamps')).toBeInTheDocument();
    expect(screen.getByText('Previous Container')).toBeInTheDocument();
    expect(screen.getByText('Word Wrap')).toBeInTheDocument();
  });

  it('should call onTailLinesChange when selecting lines', () => {
    render(<LogsFilterBar {...defaultProps} showSettings={true} />);

    const select = screen.getByRole('combobox') as HTMLSelectElement;
    fireEvent.change(select, { target: { value: '200' } });

    expect(defaultProps.onTailLinesChange).toHaveBeenCalledWith('200');
  });

  it('should call onTimestampsChange when toggling timestamps', () => {
    render(<LogsFilterBar {...defaultProps} showSettings={true} />);

    const toggleButtons = screen.getAllByRole('button');
    const timestampsToggle = toggleButtons.find(
      btn => btn.parentElement?.querySelector('label')?.textContent === 'Show Timestamps'
    );

    if (timestampsToggle) {
      fireEvent.click(timestampsToggle);
      expect(defaultProps.onTimestampsChange).toHaveBeenCalledWith(true);
    }
  });

  it('should call onPreviousChange when toggling previous', () => {
    render(<LogsFilterBar {...defaultProps} showSettings={true} />);

    const toggleButtons = screen.getAllByRole('button');
    const previousToggle = toggleButtons.find(
      btn => btn.parentElement?.querySelector('label')?.textContent === 'Previous Container'
    );

    if (previousToggle) {
      fireEvent.click(previousToggle);
      expect(defaultProps.onPreviousChange).toHaveBeenCalledWith(true);
    }
  });

  it('should call onWrapLinesChange when toggling wrap lines', () => {
    render(<LogsFilterBar {...defaultProps} showSettings={true} />);

    const toggleButtons = screen.getAllByRole('button');
    const wrapLinesToggle = toggleButtons.find(
      btn => btn.parentElement?.querySelector('label')?.textContent === 'Word Wrap'
    );

    if (wrapLinesToggle) {
      fireEvent.click(wrapLinesToggle);
      expect(defaultProps.onWrapLinesChange).toHaveBeenCalledWith(false);
    }
  });

  it('should render download button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    expect(screen.getByTitle('Download logs')).toBeInTheDocument();
  });

  it('should call onDownload when clicking download button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const downloadBtn = screen.getByTitle('Download logs');
    fireEvent.click(downloadBtn);

    expect(defaultProps.onDownload).toHaveBeenCalled();
  });

  it('should render clear button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    expect(screen.getByTitle('Clear logs')).toBeInTheDocument();
  });

  it('should call onClear when clicking clear button', () => {
    render(<LogsFilterBar {...defaultProps} />);

    const clearBtn = screen.getByTitle('Clear logs');
    fireEvent.click(clearBtn);

    expect(defaultProps.onClear).toHaveBeenCalled();
  });

  it('should show connected status', () => {
    render(<LogsFilterBar {...defaultProps} connected={true} />);

    expect(screen.getByText('Live')).toBeInTheDocument();
  });

  it('should show disconnected status', () => {
    render(<LogsFilterBar {...defaultProps} connected={false} />);

    expect(screen.getByText('Disconnected')).toBeInTheDocument();
  });

  it('should show clear button when has search term', () => {
    render(<LogsFilterBar {...defaultProps} searchTerm="test" />);

    const buttons = screen.getAllByRole('button');
    const clearBtn = buttons.find(btn => 
      btn.classList.contains('clear-btn')
    );
    expect(clearBtn).toBeInTheDocument();
  });

  it('should not show clear button when no search term', () => {
    render(<LogsFilterBar {...defaultProps} searchTerm="" />);

    const buttons = screen.getAllByRole('button');
    const clearBtn = buttons.find(btn => 
      btn.classList.contains('clear-btn')
    );
    expect(clearBtn).toBeUndefined();
  });
});
