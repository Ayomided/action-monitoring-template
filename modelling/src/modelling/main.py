import pandas as pd
import numpy as np
from pandas import Timestamp
from sklearn.ensemble import RandomForestClassifier, RandomForestRegressor
from sklearn.metrics import accuracy_score, mean_squared_error
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.preprocessing import  LabelEncoder
import tensorflow as tf
import dask.dataframe as dd
import warnings                        # To ignore any warnings
warnings.filterwarnings("ignore")

from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, LSTM, Dropout, BatchNormalization
from tensorflow.keras.callbacks import EarlyStopping, ReduceLROnPlateau
from prophet import Prophet
from prophet.serialize import model_to_json
import joblib
from utils import log

class WorkflowModelTrainer:
    def __init__(self, data_path:str) -> None:
        logger.info(f"Starting data preprocessing from file: {data_path}")
        logger.info("Loading data with Dask")
        self.data = dd.read_csv(data_path)
        logger.info(f"Loaded {len(self.data)} rows of data")
        self.X = None
        self.y = None
        self.models = {}

    def prepocess_data(self):
        logger.info("Converting datetime columns")
        self.data['StartedAt'] = dd.to_datetime(self.data['StartedAt'], utc=True)
        self.data['CompletedAt'] = dd.to_datetime(self.data['CompletedAt'], utc=True)

        logger.info("Calculating execution time")
        self.data["ExecutionTime"] = (self.data['CompletedAt'] - self.data['StartedAt']).dt.total_seconds()

        self.y_regression = self.data['ExecutionTime']
        self.y_classification = (self.data['Conclusion'] == 'failure').astype(int)

        logger.info("Encoding categorical variables")
        status_values = self.data['Status'].unique().compute()
        logger.info(f"Unique status values: {status_values}")
        status_map = {status: i for i, status in enumerate(status_values)}
        self.data['StatusEncoded'] = self.data['Status'].map(status_map)

        logger.info("Encoding Success/Failure")
        self.data['SuccessEncoded'] = (self.data['Conclusion'] == 'failure').astype(int)

        self.X = self.data[['StatusEncoded', 'ExecutionTime']]
        self.y = (self.data['Conclusion'] == 'success').astype(int)

        logger.info("Sorting data by StartedAt")
        self.sequence_data = self.data.sort_values('StartedAt')[['StatusEncoded', 'ExecutionTime', 'SuccessEncoded']]

        logger.info("Computing final DataFrame")
        self.sequence_data_pd = self.sequence_data.compute()
        logger.info(f"Columns in computed sequence data: {self.sequence_data_pd.columns}")

        # self.time_series_data = self.data.groupby('StartedAt').size().reset_index()
        # self.time_series_data.columns = ['ds', 'y']
        logger.info(f"Preprocessing complete. Final shape: {self.sequence_data.shape}")
        logger.info("Preprocesing complete")

    def prepare_sequence(self, sequence_length=10):
        logger.info(f"Preparing sequences with length: {sequence_length}")
        X = []
        y = []

        for i in range(len(self.sequence_data_pd) - sequence_length):
            X.append(self.sequence_data_pd.iloc[i:i+sequence_length][['StatusEncoded', 'ExecutionTime']].values)
            y.append(self.sequence_data_pd.iloc[i+sequence_length]['SuccessEncoded'])

        X = np.array(X)
        y = np.array(y)

        logger.info(f"Created sequences. X shape: {X.shape}, y shape: {y.shape}")

        logger.info("Normalizing ExecutionTime")
        X[:,:,1] = (X[:,:,1] - np.mean(X[:,:,1])) / np.std(X[:,:,1])

        logger.info("Splitting data into train and test sets")
        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

        logger.info(f"Train shapes - X: {X_train.shape}, y: {y_train.shape}")
        logger.info(f"Test shapes - X: {X_test.shape}, y: {y_test.shape}")

        return X_train, X_test, y_train, y_test

    # def train_random_forest(self):
    #     logger.info("Training Random Forest Model")
    #     X_train, X_test, y_train, y_test = train_test_split(self.X, self.y, test_size=0.2, random_state=42)
    #     rf_model = RandomForestClassifier(n_estimators=100, random_state=42)
    #     rf_model.fit(X_train, y_train)
    #
    #     y_pred = rf_model.predict(X_test)
    #     accuracy = accuracy_score(y_test, y_pred)
    #     # print('Random Forest Model: ')
    #     # print(f"Accuracy: {accuracy}")
    #     logger.info(f"Accuracy: {accuracy}")
    #
    #     self.models['random_forest'] = rf_model

    # def train_regression_model(self):
    #     logger.info("Training Regression")
    #     X_train, X_test, y_train, y_test = train_test_split(self.X, self.y_regression, test_size=0.2, random_state=42)
    #
    #     regression_model = RandomForestRegressor(n_estimators=100, random_state=42)
    #     regression_model.fit(X_train, y_train)
    #
    #     # Evaluate model
    #     y_pred = regression_model.predict(X_test)
    #     mse = mean_squared_error(y_test, y_pred)
    #     # print(f"Mean Squared Error: {mse}")
    #     logger.info("The Mean Squared Error: {}".format(mse))
    #
    #     # Cross-validation
    #     cv_scores = cross_val_score(regression_model, self.X, self.y_regression, cv=5)
    #     logger.info(f"Cross-validation scores: {cv_scores}")
    #     logger.info(f"Mean CV score: {np.mean(cv_scores)}")
    #     logger.info("-------------------------------")
    #     logger.info("End Regression")
    #     # print(f"Cross-validation scores: {cv_scores}")
    #     # print(f"Mean CV score: {np.mean(cv_scores)}")
    #
    #     self.models['regression'] = regression_model

    def train_rnn(self):
        logger.info("Start RNN")
        X_train, X_test, y_train, y_test = self.prepare_sequence()

        model = Sequential([
            LSTM(64, return_sequences=True, input_shape=(X_train.shape[1], X_train.shape[2])),
            BatchNormalization(),
            Dropout(0.3),
            LSTM(32),
            BatchNormalization(),
            Dropout(0.3),
            Dense(16, activation='relu'),
            Dense(1, activation='sigmoid')
        ])

        model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])

        callbacks = [
            EarlyStopping(patience=10, restore_best_weights=True),
            ReduceLROnPlateau(factor=0.5, patience=5)
        ]

        history = model.fit(
        X_train, y_train,
        validation_split=0.2,
        epochs=100,
        batch_size=64,
        callbacks=callbacks
        )

        test_loss, test_accuracy = model.evaluate(X_test, y_test)
        logger.info(f"Test accuracy: {test_accuracy:.4f} - Test loss: {test_loss:.4f}")
        logger.info("-------------------------------")
        logger.info("End RNN")
        self.models['rnn'] = model

    # def train_time_series(self):
    #     logger.info("Start Prophet")
    #     model = Prophet()
    #     model.fit(self.time_series_data)
    #
    #     future = model.make_future_dataframe(periods=30)
    #     forecast = model.predict(future)
    #
    #     logger.info("Time Series Forecasting Model (Prophet) trained successfully.")
    #     logger.info("-------------------------------")
    #     logger.info("End Prophet")
    #     self.models['prophet'] = model

    def save_models(self, base_path: str):
        joblib.dump(self.models['random_forest'], f"{base_path}_rf.joblib")
        joblib.dump(self.models['regression'], f"{base_path}_regression.joblib")
        self.models['rnn'].save(f"{base_path}_rnn.keras")
        # with open(f"{base_path}_prophet.json", "w") as fout:
        #     fout.write(model_to_json(self.models['prophet']))

    def train_all_models(self):
        logger.info("Start training")
        self.prepocess_data()
        # self.train_random_forest()
        self.train_rnn()
        # self.train_time_series()
        # self.train_regression_model()
        logger.info("-------------------------------")
        logger.info("Models Succesfully Trained")


# trainer = WorkflowModelTrainer('all_steps.csv')
logger = log(path="", file="training.logs")
trainer = WorkflowModelTrainer('/Users/davidayomide/Downloads/Dev/FINALPROJ/action-monitoring-template/modelling/MINE/all_steps.csv')
trainer.train_all_models()
trainer.save_models('workflow_models')
