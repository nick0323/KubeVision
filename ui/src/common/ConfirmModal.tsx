import React from 'react';
import { FaExclamationTriangle, FaTimes } from 'react-icons/fa';
import './ConfirmModal.css';

export interface ConfirmModalProps {
  open: boolean;
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
  icon?: React.ReactNode;
  showClose?: boolean;
  closeOnOverlay?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export const ConfirmModal: React.FC<ConfirmModalProps> = ({
  open,
  title = 'Confirm Operation',
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  danger = false,
  icon,
  showClose = true,
  closeOnOverlay = true,
  onConfirm,
  onCancel,
}) => {
  if (!open) return null;

  return (
    <div className="confirm-overlay" onClick={closeOnOverlay ? onCancel : undefined}>
      <div className="confirm-dialog" onClick={e => e.stopPropagation()}>
        <div className={`confirm-header${danger ? ' danger' : ''}`}>
          <h3 className="confirm-title">
            {icon !== undefined ? icon : danger ? <FaExclamationTriangle className="confirm-icon" /> : null}
            {title}
          </h3>
          {showClose && (
            <button className="confirm-close" onClick={onCancel}>
              <FaTimes />
            </button>
          )}
        </div>
        <div className="confirm-body">
          <p className="confirm-message">{message}</p>
        </div>
        <div className="confirm-footer">
          <button className="confirm-btn cancel" onClick={onCancel}>
            {cancelText}
          </button>
          <button
            className={`confirm-btn${danger ? ' danger' : ' primary'}`}
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
