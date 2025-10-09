-- +goose Up
-- Insert sample users
INSERT INTO users (username, email, display_name, avatar_url) VALUES
('admin', 'admin@genshinquiz.com', 'Quiz Master', 'https://example.com/avatars/admin.jpg'),
('traveler', 'traveler@example.com', 'Aether', 'https://example.com/avatars/aether.jpg'),
('outlander', 'outlander@example.com', 'Lumine', 'https://example.com/avatars/lumine.jpg');

-- Insert sample quizzes
INSERT INTO quizzes (title, description, category, difficulty, time_limit, created_by) VALUES
('Anemo Archon Quiz', 'Test your knowledge about Venti and Anemo elements', 'characters', 'easy', 180, 1),
('Weapon Mastery', 'How well do you know Genshin weapons?', 'weapons', 'medium', 300, 1),
('Artifact Expertise', 'Advanced quiz about artifact sets and stats', 'artifacts', 'hard', 600, 1);

-- Insert sample questions for Anemo Archon Quiz (quiz_id = 1)
INSERT INTO questions (quiz_id, question_text, question_type, options, correct_answer, explanation, points, order_index) VALUES
(1, 'What is the real name of the Anemo Archon?', 'multiple_choice', 
 '{"Barbatos", "Morax", "Baal", "Kusanali"}', 'Barbatos', 
 'Barbatos is the true name of the Anemo Archon, also known as Venti', 10, 1),

(1, 'Venti is a bard from which city?', 'multiple_choice',
 '{"Mondstadt", "Liyue", "Inazuma", "Sumeru"}', 'Mondstadt',
 'Venti is a bard from Mondstadt, the City of Freedom', 10, 2),

(1, 'True or False: Venti loves apples', 'true_false',
 '{"True", "False"}', 'False',
 'Venti actually dislikes apples and prefers wine instead', 5, 3);

-- Insert sample questions for Weapon Mastery Quiz (quiz_id = 2)
INSERT INTO questions (quiz_id, question_text, question_type, options, correct_answer, explanation, points, order_index) VALUES
(2, 'Which weapon type does Diluc use?', 'multiple_choice',
 '{"Sword", "Claymore", "Polearm", "Bow"}', 'Claymore',
 'Diluc wields a claymore as his primary weapon', 10, 1),

(2, 'What is the highest weapon rarity in Genshin Impact?', 'multiple_choice',
 '{"3-star", "4-star", "5-star", "6-star"}', '5-star',
 '5-star weapons are the highest rarity currently available', 15, 2);

-- Insert sample questions for Artifact Expertise Quiz (quiz_id = 3)
INSERT INTO questions (quiz_id, question_text, question_type, options, correct_answer, explanation, points, order_index) VALUES
(3, 'How many artifact pieces can a character equip?', 'multiple_choice',
 '{"3", "4", "5", "6"}', '5',
 'Characters can equip 5 artifact pieces: Flower, Feather, Sands, Goblet, and Circlet', 20, 1),

(3, 'What is the maximum artifact level?', 'fill_in_blank',
 '{}', '20',
 'Artifacts can be enhanced up to level 20', 25, 2);

-- +goose Down
DELETE FROM user_answers;
DELETE FROM quiz_attempts;
DELETE FROM questions;
DELETE FROM quizzes;
DELETE FROM users;