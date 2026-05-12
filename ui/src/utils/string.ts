export const truncateText = (text: string, maxLength: number) => {
  if (!text) return '';
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

export const capitalize = (str: string) => {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
};

export const stripAnsiCodes = (str: string): string => {
  return str.replace(/\u001b\[[0-9;]*m/g, '');
};

export const classifyLogLine = (line: string, baseClass = 'log-line'): string => {
  let className = baseClass;
  if (line.toLowerCase().includes('error')) className += ' error';
  else if (line.toLowerCase().includes('warn')) className += ' warn';
  else if (line.toLowerCase().includes('info')) className += ' info';
  return className;
};
