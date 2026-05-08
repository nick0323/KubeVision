import React, { useState, useCallback, useEffect, useRef } from 'react';
import './Dialog.css';

export interface DialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'info' | 'warning' | 'danger';
  isLoading?: boolean;
}

export const ConfirmDialog: React.FC<DialogProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = '确认',
  cancelText = '取消',
  type = 'info',
  isLoading = false,
}) => {
  const confirmButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (isOpen && confirmButtonRef.current) {
      confirmButtonRef.current.focus();
    }
  }, [isOpen]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    },
    [onClose]
  );

  if (!isOpen) return null;

  return (
    <div className="dialog-overlay" onClick={onClose} onKeyDown={handleKeyDown} role="presentation">
      <div
        className={`dialog dialog-${type}`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="dialog-title"
        aria-describedby="dialog-message"
      >
        <div className="dialog-header">
          <h3 id="dialog-title" className="dialog-title">
            {title}
          </h3>
          <button
            className="dialog-close"
            onClick={onClose}
            aria-label="关闭对话框"
            disabled={isLoading}
          >
            ×
          </button>
        </div>
        <div className="dialog-body">
          <p id="dialog-message" className="dialog-message">
            {message}
          </p>
        </div>
        <div className="dialog-footer">
          <button
            className="dialog-btn dialog-btn-cancel"
            onClick={onClose}
            disabled={isLoading}
          >
            {cancelText}
          </button>
          <button
            ref={confirmButtonRef}
            className={`dialog-btn dialog-btn-confirm dialog-btn-${type}`}
            onClick={onConfirm}
            disabled={isLoading}
          >
            {isLoading ? '处理中...' : confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export interface AlertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  message: string;
  confirmText?: string;
  type?: 'info' | 'success' | 'warning' | 'error';
}

export const AlertDialog: React.FC<AlertDialogProps> = ({
  isOpen,
  onClose,
  title,
  message,
  confirmText = '确定',
  type = 'info',
}) => {
  const confirmButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (isOpen && confirmButtonRef.current) {
      confirmButtonRef.current.focus();
    }
  }, [isOpen]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape' || e.key === 'Enter') {
        onClose();
      }
    },
    [onClose]
  );

  if (!isOpen) return null;

  return (
    <div className="dialog-overlay" onClick={onClose} onKeyDown={handleKeyDown} role="presentation">
      <div
        className={`dialog dialog-${type}`}
        onClick={(e) => e.stopPropagation()}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="alert-title"
        aria-describedby="alert-message"
      >
        <div className="dialog-header">
          <h3 id="alert-title" className="dialog-title">
            {title}
          </h3>
          <button
            className="dialog-close"
            onClick={onClose}
            aria-label="关闭对话框"
          >
            ×
          </button>
        </div>
        <div className="dialog-body">
          <p id="alert-message" className="dialog-message">
            {message}
          </p>
        </div>
        <div className="dialog-footer">
          <button
            ref={confirmButtonRef}
            className={`dialog-btn dialog-btn-confirm dialog-btn-${type}`}
            onClick={onClose}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};
