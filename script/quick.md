source venv/bin/activate
deactivate

psql -h localhost -U postgres
sudo su - postgres

python3 scripts/resets_db.py
