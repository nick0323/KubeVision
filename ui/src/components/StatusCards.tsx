import React from 'react';
import './StatusCards.css';

interface StatusCard {
  title: string;
  value: string | number;
  sub?: string;
}

interface StatusCardsProps {
  cards: StatusCard[];
}

export const StatusCards: React.FC<StatusCardsProps> = ({ cards }) => {
  return (
    <div className="status-cards">
      {cards.map((card, index) => (
        <div key={index} className="status-card">
          <div className="status-card-title">{card.title}</div>
          <div className="status-card-value">{card.value}</div>
          {card.sub && (
            <div className="status-card-sub">{card.sub}</div>
          )}
        </div>
      ))}
    </div>
  );
};

export default StatusCards;
