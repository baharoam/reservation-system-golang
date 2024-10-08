package dbrepo

import (
	"context"
	"errors"
	"log"

	"time"

	"github.com/baharoam/reservation/internal/models"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}


// InsertReservation inserts a reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date,
			end_date, room_id, created_at, updated_at) 
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) GetFirstRoom(res models.Room) error {
	var roomID int
	var roomName string
	stmt := `SELECT id, room_name FROM rooms LIMIT 1`
    err := m.DB.QueryRow(stmt).Scan(&roomID, &roomName)
    if err != nil {
        return err
    }
    log.Println("Successfully connected to the database. Fetched room with ID This is from postgres file:", roomID, roomName)
    return nil
}


// InsertRoomRestriction inserts a room restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id,	
			created_at, updated_at, restriction_id) 
			values
			($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) IsEmailUnique(email string) bool {
    log.Println("Checking email:", email)
    
    var dbEmail string
    query := `SELECT email FROM reservations WHERE email = $1`
    err := m.DB.QueryRow(query, email).Scan(&dbEmail)
    
    if err != nil {
        if err == pgx.ErrNoRows {
            // Email not found in the database
            return false
        }
        // Some other error occurred
        log.Println("Email is unique", err)
        return false
    }
    
    // Email was found in the database
    log.Println("Email found in database:", dbEmail)
    return true
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query :=
	`select 
		count(id)
	from 
		room_restrictions
	where
		room_id = $1
		and $2 < end_date and $3 > start_date;`
	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query :=
	`select 
		r.id, r.room_name
	from
		rooms r
	where r.id not in
	(select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date);`
	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next(){
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var room models.Room

	query := `
	select id, room_name, created_at, updated_at from rooms where id = $1
	`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil{
		return room, err

	}
	return room, nil
    }

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `select id, first_name, last_name, email, password, access_level, created_at, updated_at
		from users where id = $1
		`
		var u models.User

		row := m.DB.QueryRowContext(ctx, query, id)

		err := row.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Password,
			&u.AccessLevel,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil{
			return u, err
		}
		return u, nil
	}


func (m *postgresDBRepo) UpdateUser (u models.User) error{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `update users set first_name =$1, last_name =$2, email =$3, access_level =$4, updated_at= $5
		`
		_,err := m.DB.ExecContext(ctx, query,
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Password,
			&u.AccessLevel,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil{
			return err
		}
		return nil
	}


func (m *postgresDBRepo) Authenticate (email, testPassword string) (int, string, error){
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var id int
		var hashedPassword string

		row := m.DB.QueryRowContext(ctx, "select id, password from users where email =$1", email)

		err:= row.Scan(&id, &hashedPassword)

		if err != nil {
			return id, "", err
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, "", errors.New("incorrec password")
		} else if err != nil{
			return 0, "", err
		}
		return id, hashedPassword, nil
	}


func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
	select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
	r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from reservations r
	left join rooms rm on (r.room_id = rm.id)
	order by r.start_date asc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next(){
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}

func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
	select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
	r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from reservations r
	left join rooms rm on (r.room_id = rm.id)
	where processed = 0 
	order by r.start_date asc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next(){
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}

func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
	select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
	r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from reservations r
	left join rooms rm on (r.room_id = rm.id)
	where r.id = $1
	`

	rows := m.DB.QueryRowContext(ctx, query, id)
	err := rows.Scan(
			&res.ID,
			&res.FirstName,
			&res.LastName,
			&res.Email,
			&res.Phone,
			&res.StartDate,
			&res.EndDate,
			&res.RoomID,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Processed,
			&res.Room.ID,
			&res.Room.RoomName,
		)

		if err != nil {
			return res, err
		}
		return res, nil
	}


func (m *postgresDBRepo) UpdateReservation (u models.Reservation) error{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5 where id = $6
		`
		_,err := m.DB.ExecContext(ctx, query,
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Phone,
			time.Now(),
			u.ID,
		)
		if err != nil{
			return err
		}
		return nil
	}

func (m *postgresDBRepo) DeleteReservation (id int) error{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `delete from reservations where id = $1
		`
		_,err := m.DB.ExecContext(ctx, query,id)
		if err != nil{
			return err
		}
		return nil
	}

func (m *postgresDBRepo) UpdateProcessedForReservation (id, processed int) error{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `update reservations set processed = $1 where id = $2
		`
		_,err := m.DB.ExecContext(ctx, query, processed, id)
		if err != nil{
			return err
		}
		return nil
	}