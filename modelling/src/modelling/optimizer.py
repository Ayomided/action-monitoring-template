import pandas as pd
import numpy as np
import joblib

class WorkflowOptimizer:
    def __init__(self, data_path, rf_model_path, classification_model_path):
        self.data = pd.read_csv(data_path)
        self.regression_model = joblib.load(rf_model_path)
        self.classification_model = joblib.load(classification_model_path)
        self.data['Failed'] = (self.data['Conclusion'] == 'failure').astype(int)
        self.prepare_features()

    def prepare_features(self):
        feature_columns = ['StepName', 'ExecutionTime', 'Status', 'Attempt', 'JobName']
        self.features = self.data[feature_columns].copy()

        categorical_columns = ['StepName', 'Status', 'JobName']
        self.features = pd.get_dummies(self.features, columns=categorical_columns)

        self.features['Attempt'] = pd.to_numeric(self.features['Attempt'], errors='coerce')

        # Ensure all features used during model training are present
        model_features = self.regression_model.feature_names_in_
        for feature in model_features:
            if feature not in self.features.columns:
                self.features[feature] = 0

        # Reorder columns to match the model's expected feature order
        self.features = self.features.reindex(columns=model_features, fill_value=0)

    def identify_bottlenecks(self):
        feature_importance = self.regression_model.feature_importances_
        features = self.regression_model.feature_names_in_

        importance_df = pd.DataFrame({'feature': features, 'importance': feature_importance})
        bottlenecks = importance_df.sort_values('importance', ascending=False).head(5)
        return bottlenecks

    def suggest_parallelization(self):
        # Simplified logic: suggest parallelization if multiple important features
        bottlenecks = self.identify_bottlenecks()
        step_bottlenecks = [feat for feat in bottlenecks['feature'] if feat.startswith('StepName_')]
        if len(step_bottlenecks) > 2:
            steps_to_parallelize = [feat.replace('StepName_', '') for feat in step_bottlenecks[:3]]
            return "Consider parallelizing steps related to: " + ", ".join(steps_to_parallelize)
        return "No significant parallelization opportunities identified."

    def recommend_caching(self):
        # Simplified logic: recommend caching for features with high importance
        bottlenecks = self.identify_bottlenecks()
        cache_candidates = bottlenecks[bottlenecks['importance'] > 0.1]['feature'].tolist()
        cache_steps = [feat.replace('StepName_', '') for feat in cache_candidates if feat.startswith('StepName_')]
        if cache_steps:
            return "Consider implementing caching for: " + ", ".join(cache_steps)
        return "No significant caching opportunities identified."

    def resource_allocation_advice(self):
        avg_exec_time = np.mean(self.data['ExecutionTime'])
        if avg_exec_time > 300:  # If average execution time is over 5 minutes
            return "Consider upgrading to a higher-capacity runner or optimizing resource-intensive steps."
        return "Current resource allocation appears sufficient."

    def failure_prevention_strategies(self):
        failure_rate = np.mean(self.data['Failed'])
        if failure_rate > 0.1:  # If failure rate is over 10%
            high_failure_steps = self.data[self.data['Failed'] == 1]['StepName'].value_counts().head(3)
            step_advice = ", ".join([f"{step} ({count} failures)" for step, count in high_failure_steps.items()])
            return f"High failure rate detected. Focus on these steps: {step_advice}. Review error logs and consider implementing retry mechanisms for flaky tests."
        return "Current failure rate is within acceptable limits."

    def generate_optimization_report(self):
        report = """
        Workflow Optimization Recommendations:

        1. Bottleneck Analysis:
        {bottlenecks}

        2. Parallelization Opportunities:
        {parallelization}

        3. Caching Recommendations:
        {caching}

        4. Resource Allocation Advice:
        {resources}

        5. Failure Prevention Strategies:
        {failure_prevention}
        """

        return report.format(
            bottlenecks=self.identify_bottlenecks().to_string(),
            parallelization=self.suggest_parallelization(),
            caching=self.recommend_caching(),
            resources=self.resource_allocation_advice(),
            failure_prevention=self.failure_prevention_strategies()
        )

# Usage
optimizer = WorkflowOptimizer('workflow_data.csv', 'workflow_models_rf.joblib', 'workflow_models_regression.joblib')
optimization_report = optimizer.generate_optimization_report()
print(optimization_report)
