import pandas as pd
import numpy as np
from pandas import Timestamp
from sklearn.ensemble import RandomForestClassifier, RandomForestRegressor
from sklearn.metrics import accuracy_score, mean_squared_error
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.preprocessing import  LabelEncoder
import tensorflow as tf
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, LSTM, Input
from prophet import Prophet
from prophet.serialize import model_to_json
import joblib
from utils import parse_time



class WorkflowModelTrainer:
    def __init__(self, data_path:str) -> None:
        self.data = pd.read_csv(data_path)
        self.X = None
        self.y = None
        self.models = {}

    def prepocess_data(self):
        self.data['StartedAt'] = self.data['StartedAt'].apply(parse_time)
        self.data['CompletedAt'] = self.data['CompletedAt'].apply(parse_time)

        self.data["ExecutionTime"] = (self.data['CompletedAt'] - self.data['StartedAt']).dt.total_seconds()

        self.y_regression = self.data['ExecutionTime']
        self.y_classification = (self.data['Conclusion'] == 'failure').astype(int)

        le = LabelEncoder()
        self.data['StatusEncoded'] = le.fit_transform(self.data['Status'])

        self.X = self.data[['StatusEncoded', 'ExecutionTime']]
        self.y = (self.data['Conclusion'] == 'success').astype(int)

        self.data.to_csv('workflow_data.csv', index=False)

        self.sequence_data = self.data.sort_values('StartedAt')[['StatusEncoded', 'ExecutionTime', 'Conclusion']]
        self.sequence_data['SuccessEncoded']= (self.sequence_data['Conclusion'] == 'success').astype(int)

        self.time_series_data = self.data.groupby('StartedAt').size().reset_index(name='count')
        self.time_series_data.columns = ['ds', 'y']

    def train_random_forest(self):
        X_train, X_test, y_train, y_test = train_test_split(self.X, self.y, test_size=0.2, random_state=42)
        rf_model = RandomForestClassifier(n_estimators=100, random_state=42)
        rf_model.fit(X_train, y_train)

        y_pred = rf_model.predict(X_test)
        accuracy = accuracy_score(y_test, y_pred)
        print('Random Forest Model: ')
        print(f"Accuracy: {accuracy}")

        self.models['random_forest'] = rf_model

    def train_regression_model(self):
        X_train, X_test, y_train, y_test = train_test_split(self.X, self.y_regression, test_size=0.2, random_state=42)

        regression_model = RandomForestRegressor(n_estimators=100, random_state=42)
        regression_model.fit(X_train, y_train)

        # Evaluate model
        y_pred = regression_model.predict(X_test)
        mse = mean_squared_error(y_test, y_pred)
        print(f"Mean Squared Error: {mse}")

        # Cross-validation
        cv_scores = cross_val_score(regression_model, self.X, self.y_regression, cv=5)
        print(f"Cross-validation scores: {cv_scores}")
        print(f"Mean CV score: {np.mean(cv_scores)}")

        self.models['regression'] = regression_model

    def train_rnn(self):
        sequence_length = 10
        X = []
        y = []
        for i in range(len(self.sequence_data)- sequence_length):
            X.append(self.sequence_data.iloc[i:i+sequence_length][['StatusEncoded', 'ExecutionTime']].values)
            y.append(self.sequence_data.iloc[i+sequence_length]['SuccessEncoded'])
        X = np.array(X)
        y = np.array(y)

        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

        model = Sequential([
            Input(shape=(sequence_length, 2)),
            LSTM(64, return_sequences=True),
            LSTM(32),
            Dense(1, activation='sigmoid')
        ])
        model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])

        history = model.fit(X_train, y_train, epochs=50, batch_size=32, validation_split=0.2, verbose=1)

        _, accuracy = model.evaluate(X_test, y_test, verbose=1)
        print("RNN Model:")
        print(f"Accuracy: {accuracy}")
        self.models['rnn'] = model

    def train_time_series(self):
        model = Prophet()
        model.fit(self.time_series_data)

        future = model.make_future_dataframe(periods=30)
        forecast = model.predict(future)

        print("Time Series Forecasting Model (Prophet) trained successfully.")
        self.models['prophet'] = model

    def save_models(self, base_path: str):
        joblib.dump(self.models['random_forest'], f"{base_path}_rf.joblib")
        joblib.dump(self.models['regression'], f"{base_path}_regression.joblib")
        self.models['rnn'].save(f"{base_path}_rnn.keras")
        with open(f"{base_path}_prophet.json", "w") as fout:
            fout.write(model_to_json(self.models['prophet']))

    def train_all_models(self):
        self.prepocess_data()
        self.train_random_forest()
        self.train_rnn()
        self.train_time_series()
        self.train_regression_model()


trainer = WorkflowModelTrainer('all_steps.csv')
trainer.train_all_models()
trainer.save_models('workflow_models')
