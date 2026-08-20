package storage

import (
	"context"
	"database/sql"
)

func (s *Store) ListIPMINodes(ctx context.Context) ([]IPMINode, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,COALESCE(client_id,0),name,address,username,password,interface,vendor,created_at,updated_at FROM ipmi_nodes ORDER BY name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IPMINode
	for rows.Next() {
		var n IPMINode
		if err := rows.Scan(&n.ID, &n.ClientID, &n.Name, &n.Address, &n.Username, &n.Password, &n.Interface, &n.Vendor, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) GetIPMINode(ctx context.Context, id int64) (IPMINode, error) {
	var n IPMINode
	err := s.db.QueryRowContext(ctx, `SELECT id,COALESCE(client_id,0),name,address,username,password,interface,vendor,created_at,updated_at FROM ipmi_nodes WHERE id=?`, id).Scan(&n.ID, &n.ClientID, &n.Name, &n.Address, &n.Username, &n.Password, &n.Interface, &n.Vendor, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func (s *Store) SaveIPMINode(ctx context.Context, n IPMINode) (IPMINode, error) {
	now := Now()
	if n.ID == 0 {
		res, err := s.db.ExecContext(ctx, `INSERT INTO ipmi_nodes(client_id,name,address,username,password,interface,vendor,created_at,updated_at) VALUES(NULLIF(?,0),?,?,?,?,?,?,?,?)`, n.ClientID, n.Name, n.Address, n.Username, n.Password, n.Interface, n.Vendor, now, now)
		if err != nil {
			return n, err
		}
		n.ID, _ = res.LastInsertId()
	} else {
		if n.Password == "" {
			_, err := s.db.ExecContext(ctx, `UPDATE ipmi_nodes SET client_id=NULLIF(?,0),name=?,address=?,username=?,interface=?,vendor=?,updated_at=? WHERE id=?`, n.ClientID, n.Name, n.Address, n.Username, n.Interface, n.Vendor, now, n.ID)
			if err != nil {
				return n, err
			}
		} else {
			_, err := s.db.ExecContext(ctx, `UPDATE ipmi_nodes SET client_id=NULLIF(?,0),name=?,address=?,username=?,password=?,interface=?,vendor=?,updated_at=? WHERE id=?`, n.ClientID, n.Name, n.Address, n.Username, n.Password, n.Interface, n.Vendor, now, n.ID)
			if err != nil {
				return n, err
			}
		}
	}
	return s.GetIPMINode(ctx, n.ID)
}

func (s *Store) DeleteIPMINode(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM ipmi_nodes WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
