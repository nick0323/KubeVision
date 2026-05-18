import React, { useState, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { FiGlobe, FiKey, FiUsers, FiLogOut } from 'react-icons/fi';
import { notification } from '../common/NotificationContext';
import apiClient from '../utils/apiClient';

interface SettingsModalProps {
  open: boolean;
  onClose: () => void;
  onManageClusters: () => void;
  onLogout: () => void;
}

const SettingsModal: React.FC<SettingsModalProps> = ({ open, onClose, onManageClusters, onLogout }) => {
  const [showPassword, setShowPassword] = useState(false);
  const [form, setForm] = useState({ oldPassword: '', newPassword: '', confirmPassword: '' });
  const [changing, setChanging] = useState(false);

  const resetForm = useCallback(() => {
    setForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
  }, []);

  const handlePasswordChange = useCallback(async () => {
    if (!form.oldPassword || !form.newPassword) {
      notification.warning('Please fill in all fields');
      return;
    }
    if (form.newPassword !== form.confirmPassword) {
      notification.warning('New passwords do not match');
      return;
    }
    if (form.newPassword.length < 8) {
      notification.warning('Password must be at least 8 characters');
      return;
    }
    setChanging(true);
    try {
      await apiClient.post('/api/v1/admin/password/change', {
        oldPassword: form.oldPassword,
        newPassword: form.newPassword,
      });
      notification.success('Password changed successfully');
      setShowPassword(false);
      resetForm();
    } catch (err) {
      notification.error(`Failed to change password: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setChanging(false);
    }
  }, [form, resetForm]);

  if (!open) return null;

  return createPortal(
    <>
      <div className="settings-overlay" onClick={onClose}>
        <div className="settings-page" onClick={e => e.stopPropagation()}>
          <div className="settings-page-header">
            <h2>Settings</h2>
            <button className="settings-close-btn" onClick={onClose}>✕</button>
          </div>

          <div className="settings-page-body">
            <div className="settings-section-label">Administration</div>
            <div className="settings-action-list">
              <div className="settings-action" onClick={onManageClusters}>
                <FiGlobe size={18} />
                <div className="settings-action-text">
                  <span className="settings-action-title">Manage Clusters</span>
                  <span className="settings-action-desc">Add, edit, or remove cluster connections</span>
                </div>
              </div>
              <div className="settings-action" onClick={() => setShowPassword(true)}>
                <FiKey size={18} />
                <div className="settings-action-text">
                  <span className="settings-action-title">Change Password</span>
                  <span className="settings-action-desc">Update your login password</span>
                </div>
              </div>
              <div className="settings-action disabled" title="Coming soon">
                <FiUsers size={18} />
                <div className="settings-action-text">
                  <span className="settings-action-title">User Management</span>
                  <span className="settings-action-desc">Manage user accounts and permissions</span>
                </div>
              </div>
            </div>

            <hr className="settings-divider" />

            <button className="settings-logout-btn" onClick={onLogout}>
              <FiLogOut size={18} />
              <span>Log Out</span>
            </button>
          </div>
        </div>
      </div>

      {showPassword && (
        <div className="password-modal-overlay" onClick={() => { setShowPassword(false); resetForm(); }}>
          <div className="password-modal" onClick={e => e.stopPropagation()}>
            <h3>Change Password</h3>
            <div className="form-field">
              <label>Current Password</label>
              <input
                type="password"
                value={form.oldPassword}
                onChange={e => setForm(p => ({ ...p, oldPassword: e.target.value }))}
                placeholder="Enter current password"
              />
            </div>
            <div className="form-field">
              <label>New Password</label>
              <input
                type="password"
                value={form.newPassword}
                onChange={e => setForm(p => ({ ...p, newPassword: e.target.value }))}
                placeholder="Enter new password (min 8 chars)"
              />
            </div>
            <div className="form-field">
              <label>Confirm New Password</label>
              <input
                type="password"
                value={form.confirmPassword}
                onChange={e => setForm(p => ({ ...p, confirmPassword: e.target.value }))}
                placeholder="Confirm new password"
              />
            </div>
            <div className="form-buttons">
              <button className="password-cancel-btn" onClick={() => { setShowPassword(false); resetForm(); }}>
                Cancel
              </button>
              <button className="create-resource-btn" onClick={handlePasswordChange} disabled={changing}>
                {changing ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>,
    document.body
  );
};

export default React.memo(SettingsModal);
