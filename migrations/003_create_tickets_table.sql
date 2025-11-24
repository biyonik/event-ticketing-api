-- Create tickets table
CREATE TABLE IF NOT EXISTS tickets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    seat_id BIGINT NULL,
    section_id BIGINT NOT NULL,
    ticket_number VARCHAR(50) NOT NULL UNIQUE,
    ticket_type VARCHAR(50) NOT NULL DEFAULT 'standard', -- standard, vip, early_bird, season
    status VARCHAR(50) NOT NULL DEFAULT 'reserved', -- reserved, sold, used, cancelled, expired
    price DECIMAL(10, 2) NOT NULL,
    qr_code_data TEXT NOT NULL,
    qr_code_image LONGBLOB,
    verification_code VARCHAR(10) NOT NULL,
    reservation_expiry TIMESTAMP NULL,
    purchased_at TIMESTAMP NULL,
    used_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE RESTRICT,
    FOREIGN KEY (seat_id) REFERENCES seats(id) ON DELETE RESTRICT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE RESTRICT,
    INDEX idx_event_id (event_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_ticket_number (ticket_number),
    INDEX idx_reservation_expiry (reservation_expiry)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
