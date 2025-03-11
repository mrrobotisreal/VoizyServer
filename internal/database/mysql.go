package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitMySQL() error {
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/voizy?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DBU"), os.Getenv("DBP"))
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open error: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("db.Ping error: %w", err)
	}

	if err := createTables(); err != nil {
		fmt.Println("createTables error occurred: ", err)
		return err
	}

	fmt.Println("MySQL connected and schema ensured.")
	return nil
}

func createTables() error {
	// Users and profiles
	apiKeysTable := `
	CREATE TABLE IF NOT EXISTS api_keys (
		api_key_id		  BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id			  BIGINT NOT NULL,
		api_key			  VARCHAR(255) NOT NULL UNIQUE,
		created_at		  DATETIME	   NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at		  DATETIME	   NOT NULL,
	    last_used_at	  DATETIME	   NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at		  DATETIME	   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);
	`

	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		user_id           BIGINT AUTO_INCREMENT PRIMARY KEY,
		api_key			  VARCHAR(255) NOT NULL UNIQUE,
		email             VARCHAR(255) NOT NULL UNIQUE,
	    salt			  VARCHAR(255) NOT NULL,
		password_hash     VARCHAR(255) NOT NULL,
		username          VARCHAR(50)  NOT NULL UNIQUE,
		created_at        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

	userProfilesTable := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		profile_id        BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id           BIGINT NOT NULL,
		first_name        VARCHAR(100),
		last_name         VARCHAR(100),
	    preferred_name    VARCHAR(100),
		birth_date        DATE,
		city_of_residence VARCHAR(255),
		place_of_work     VARCHAR(255),
		date_joined       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	userSchoolsTable := `
	CREATE TABLE IF NOT EXISTS user_schools (
		user_school_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id           BIGINT NOT NULL,
		school_name       VARCHAR(255) NOT NULL,
		start_year        INT,
		end_year          INT,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	interestsTable := `
	CREATE TABLE IF NOT EXISTS interests (
		interest_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name        VARCHAR(255) NOT NULL UNIQUE
	);`

	userInterestsTable := `
	CREATE TABLE IF NOT EXISTS user_interests (
		user_interest_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id          BIGINT NOT NULL,
		interest_id      BIGINT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
		FOREIGN KEY (interest_id) REFERENCES interests(interest_id) ON DELETE CASCADE
	);`

	userSocialLinksTable := `
	CREATE TABLE IF NOT EXISTS user_social_links (
		link_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id   BIGINT NOT NULL,
		platform  VARCHAR(100) NOT NULL,
		url       VARCHAR(255) NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	userImagesTable := `
	CREATE TABLE IF NOT EXISTS user_images (
		user_image_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id       BIGINT NOT NULL,
		image_url     VARCHAR(255) NOT NULL,
		is_profile_pic BOOLEAN NOT NULL DEFAULT 0,
		uploaded_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	// Friendships
	friendshipsTable := `
	CREATE TABLE IF NOT EXISTS friendships (
		friendship_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id       BIGINT NOT NULL,
		friend_id     BIGINT NOT NULL,
		status        ENUM('pending','accepted','blocked') NOT NULL DEFAULT 'pending',
		created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
		FOREIGN KEY (friend_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	// Groups
	groupsTable := `
	CREATE TABLE IF NOT EXISTS groups_table (
		group_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
		name        VARCHAR(255) NOT NULL,
		description TEXT,
		privacy     ENUM('public','private','closed') NOT NULL DEFAULT 'public',
		creator_id  BIGINT NOT NULL,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	groupMembersTable := `
	CREATE TABLE IF NOT EXISTS group_members (
		group_member_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		group_id        BIGINT NOT NULL,
		user_id         BIGINT NOT NULL,
		role            ENUM('member','moderator','admin') NOT NULL DEFAULT 'member',
		joined_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (group_id) REFERENCES groups_table(group_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	// Posts
	postsTable := `
	CREATE TABLE IF NOT EXISTS posts (
		post_id            BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id            BIGINT NOT NULL,
		to_user_id		   BIGINT NOT NULL DEFAULT -1,
		original_post_id   BIGINT NULL DEFAULT NULL,
		impressions		   BIGINT NOT NULL DEFAULT 0,
		views			   BIGINT NOT NULL DEFAULT 0,
		content_text       TEXT,
		created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		location_name      VARCHAR(255),
		location_lat       DECIMAL(9,6),
		location_lng       DECIMAL(9,6),
		is_poll            BOOLEAN NOT NULL DEFAULT 0,
		poll_question      VARCHAR(255),
		poll_duration_type ENUM('hours','days','weeks') DEFAULT 'days',
		poll_duration_length INT DEFAULT 1,
		poll_end_datetime  DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
	    FOREIGN KEY (original_post_id) REFERENCES posts(post_id) ON DELETE SET NULL
	);`

	pollOptionsTable := `
	CREATE TABLE IF NOT EXISTS poll_options (
		poll_option_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id        BIGINT NOT NULL,
		option_text    VARCHAR(255) NOT NULL,
		vote_count     INT DEFAULT 0,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
	);`

	pollVotesTable := `
	CREATE TABLE IF NOT EXISTS poll_votes (
		poll_vote_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id        BIGINT NOT NULL,
		poll_option_id BIGINT NOT NULL,
		user_id        BIGINT NOT NULL,
		voted_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (poll_option_id) REFERENCES poll_options(poll_option_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	hashtagsTable := `
	CREATE TABLE IF NOT EXISTS hashtags (
		hashtag_id  BIGINT AUTO_INCREMENT PRIMARY KEY,
		tag         VARCHAR(255) NOT NULL UNIQUE,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	postHashtagsTable := `
	CREATE TABLE IF NOT EXISTS post_hashtags (
		post_hashtag_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id         BIGINT NOT NULL,
		hashtag_id      BIGINT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (hashtag_id) REFERENCES hashtags(hashtag_id) ON DELETE CASCADE
	);`

	postReactionsTable := `
	CREATE TABLE IF NOT EXISTS post_reactions (
		post_reaction_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id          BIGINT NOT NULL,
		user_id          BIGINT NOT NULL,
		reaction_type    ENUM('like','love','laugh','congratulate','shocked','sad','angry') NOT NULL,
		reacted_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	commentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
		comment_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id      BIGINT NOT NULL,
		user_id      BIGINT NOT NULL,
		content_text TEXT NOT NULL,
		created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	commentReactionsTable := `
	CREATE TABLE IF NOT EXISTS comment_reactions (
		comment_reaction_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		comment_id          BIGINT NOT NULL,
		user_id             BIGINT NOT NULL,
		reaction_type       ENUM('like','love','laugh','congratulate','shocked','sad','angry') NOT NULL,
		reacted_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	postSharesTable := `
	CREATE TABLE IF NOT EXISTS post_shares (
		share_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id    BIGINT NOT NULL,
		user_id    BIGINT NOT NULL,
		shared_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	postMediaTable := `
	CREATE TABLE IF NOT EXISTS post_media (
		media_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
		post_id     BIGINT NOT NULL,
		media_url   VARCHAR(255) NOT NULL,
		media_type  ENUM('image','video') NOT NULL,
		uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
	);`

	// Messages and Chat
	conversationsTable := `
	CREATE TABLE IF NOT EXISTS conversations (
		conversation_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		conversation_name VARCHAR(255),
		is_group_chat   BOOLEAN NOT NULL DEFAULT 0,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	conversationMembersTable := `
	CREATE TABLE IF NOT EXISTS conversation_members (
		conv_member_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
		conversation_id  BIGINT NOT NULL,
		user_id          BIGINT NOT NULL,
		joined_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		message_id       BIGINT AUTO_INCREMENT PRIMARY KEY,
		conversation_id  BIGINT NOT NULL,
		sender_id        BIGINT NOT NULL,
		content_text     TEXT,
		sent_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id) ON DELETE CASCADE,
		FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	messageRecipientsTable := `
	CREATE TABLE IF NOT EXISTS message_recipients (
		msg_recipient_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		message_id       BIGINT NOT NULL,
		recipient_id     BIGINT NOT NULL,
		is_read          BOOLEAN NOT NULL DEFAULT 0,
		read_at          DATETIME,
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
		FOREIGN KEY (recipient_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	messageAttachmentsTable := `
	CREATE TABLE IF NOT EXISTS message_attachments (
		attachment_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		message_id    BIGINT NOT NULL,
		file_url      VARCHAR(255) NOT NULL,
		file_type     ENUM('image','video','doc') DEFAULT 'image',
		uploaded_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE
	);`

	messageReactionsTable := `
	CREATE TABLE IF NOT EXISTS message_reactions (
		message_reaction_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		message_id          BIGINT NOT NULL,
		user_id             BIGINT NOT NULL,
		reaction_type       ENUM('like','love','laugh','congratulate','shocked','sad','angry') NOT NULL,
		reacted_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	// Analytics
	analyticsEventsTable := `
	CREATE TABLE IF NOT EXISTS analytics_events (
		event_id      BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id       BIGINT NOT NULL,
		event_type    VARCHAR(100) NOT NULL,
		object_type   VARCHAR(100),
		object_id     BIGINT,
		event_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata      JSON,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(apiKeysTable); err != nil {
		return err
	}
	if _, err := DB.Exec(usersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userProfilesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userSchoolsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(interestsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userInterestsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userSocialLinksTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userImagesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(friendshipsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(groupsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(groupMembersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(pollOptionsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(pollVotesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(hashtagsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postHashtagsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postReactionsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(commentsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(commentReactionsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postSharesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postMediaTable); err != nil {
		return err
	}
	if _, err := DB.Exec(conversationsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(conversationMembersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(messagesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(messageRecipientsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(messageAttachmentsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(messageReactionsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(analyticsEventsTable); err != nil {
		return err
	}

	return nil
}
