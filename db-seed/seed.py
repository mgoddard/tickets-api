from sqlalchemy import create_engine, Column, String, ForeignKey, Numeric, DECIMAL
from sqlalchemy.orm import declarative_base, sessionmaker
import uuid
from faker import Faker
import random
from tqdm import tqdm
import argparse

fake = Faker()
Base = declarative_base()


class User(Base):
    __tablename__ = 'users'
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String)


class Event(Base):
    __tablename__ = 'events'
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String)
    type = Column(String)
    status = Column(String)


class Purchase(Base):
    __tablename__ = 'purchases'
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String, ForeignKey('users.id'))
    event_id = Column(String, ForeignKey('events.id'))
    status = Column(String)


class Payment(Base):
    __tablename__ = 'payments'
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    purchase_id = Column(String, ForeignKey('purchases.id'))
    amount = Column(DECIMAL(10, 2))
    status = Column(String)


def create_fake_data(session, num_users, num_purchases, num_cancellations, num_payments):
    user_ids = []
    for _ in tqdm(range(num_users), desc='Creating Users'):
        user = User(name=fake.name())
        session.add(user)
        session.commit()
        user_ids.append(user.id)

    event_ids = []
    for _ in tqdm(range(num_purchases), desc='Creating Events'):
        event = Event(name=fake.sentence(nb_words=4), type=random.choice(['concert', 'opera', 'theater']),
                      status='scheduled')
        session.add(event)
        session.commit()
        event_ids.append(event.id)

    purchase_ids = []
    for _ in tqdm(range(num_purchases), desc='Creating Purchases'):
        purchase = Purchase(user_id=random.choice(user_ids), event_id=random.choice(event_ids), status='confirmed')
        session.add(purchase)
        session.commit()
        purchase_ids.append(purchase.id)

    for _ in tqdm(range(num_cancellations), desc='Creating Cancellations'):
        purchase_id = random.choice(purchase_ids)
        purchase = session.query(Purchase).filter_by(id=purchase_id).first()
        purchase.status = 'cancelled'
        session.commit()

    for _ in tqdm(range(num_payments), desc='Creating Payments'):
        purchase_id = random.choice(purchase_ids)
        payment = Payment(purchase_id=purchase_id, amount=random.uniform(20, 200), status='successful')
        session.add(payment)
        session.commit()


def main(args):
    engine = create_engine('cockroachdb://root@192.168.86.74:26257/tickets')
    Base.metadata.create_all(engine)

    Session = sessionmaker(bind=engine)
    session = Session()

    create_fake_data(session, args.num_users, args.num_purchases, args.num_cancellations, args.num_payments)

    session.close()
    print("Data generation complete!")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Generate fake data for ticket purchasing service.')
    parser.add_argument('--num_users', type=int, default=1000, help='Number of users to generate')
    parser.add_argument('--num_purchases', type=int, default=5000, help='Number of purchases to generate')
    parser.add_argument('--num_cancellations', type=int, default=1000, help='Number of cancellations to generate')
    parser.add_argument('--num_payments', type=int, default=5000, help='Number of payments to generate')

    args = parser.parse_args()
    main(args)
