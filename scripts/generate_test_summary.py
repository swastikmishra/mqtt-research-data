#!/usr/bin/env python3
"""
Script to read all test-cases JSON files and create CSV files with test configuration and results.
For each test-case category, it creates a CSV file combining the test parameters with the aggregated results.
"""

import os
import json
import csv
import glob
from pathlib import Path

def find_latest_requests_file(data_dir, test_code):
    """Find the most recent requests CSV file for a given test code."""
    pattern = os.path.join(data_dir, test_code, "requests_*.csv")
    files = glob.glob(pattern)
    if not files:
        return None
    # Return the most recent file (highest timestamp)
    return max(files, key=os.path.getctime)

def get_aggregated_results(csv_file_path):
    """Extract the aggregated results row from the requests CSV file."""
    if not csv_file_path or not os.path.exists(csv_file_path):
        return None
    
    try:
        with open(csv_file_path, 'r', encoding='utf-8') as file:
            lines = file.readlines()
            
        # Find the aggregated row (last non-empty line that starts with ",Aggregated")
        for line in reversed(lines):
            line = line.strip()
            if line and line.startswith(',Aggregated'):
                # Parse the CSV line
                parts = line.split(',')
                if len(parts) >= 22:  # Ensure we have enough columns
                    return {
                        'Total_Requests': parts[2],
                        'Failure_Count': parts[3],
                        'Median_Response_Time': parts[4],
                        'Average_Response_Time': parts[5],
                        'Min_Response_Time': parts[6],
                        'Max_Response_Time': parts[7],
                        'Average_Content_Size': parts[8],
                        'Requests_per_sec': parts[9],
                        'Failures_per_sec': parts[10],
                        'Response_Time_50%': parts[11],
                        'Response_Time_66%': parts[12],
                        'Response_Time_75%': parts[13],
                        'Response_Time_80%': parts[14],
                        'Response_Time_90%': parts[15],
                        'Response_Time_95%': parts[16],
                        'Response_Time_98%': parts[17],
                        'Response_Time_99%': parts[18],
                        'Response_Time_99.9%': parts[19],
                        'Response_Time_99.99%': parts[20],
                        'Response_Time_100%': parts[21]
                    }
    except Exception as e:
        print(f"Error reading {csv_file_path}: {e}")
    
    return None

def process_test_cases():
    """Process all test-cases and create summary CSV files."""
    base_dir = "/app"
    test_cases_dir = os.path.join(base_dir, "test-cases")
    data_dir = os.path.join(base_dir, "data")

    print(data_dir)
    print(test_cases_dir)
    
    # Define the mapping from directory names to categories
    category_mapping = {
        "00 Single": "SINGLE",
        "01 Horizontal": "HORIZONTAL", 
        "02 Vertical": "VERTICAL"
    }
    
    # Process each category
    for category_dir in os.listdir(test_cases_dir):
        category_path = os.path.join(test_cases_dir, category_dir)
        if not os.path.isdir(category_path):
            continue
            
        category_name = category_mapping.get(category_dir, category_dir)
        print(f"Processing category: {category_name}")
        
        # Collect all test cases for this category
        all_test_cases = []
        
        # Process each subcategory (broker configuration)
        for subcat_dir in os.listdir(category_path):
            subcat_path = os.path.join(category_path, subcat_dir)
            if not os.path.isdir(subcat_path):
                continue
                
            tests_json_path = os.path.join(subcat_path, "tests.json")
            if not os.path.exists(tests_json_path):
                continue
                
            try:
                with open(tests_json_path, 'r', encoding='utf-8') as f:
                    test_cases = json.load(f)
                    
                for test_case in test_cases:
                    # Find corresponding data directory
                    data_category_path = os.path.join(data_dir, category_dir, subcat_dir)
                    
                    # Get aggregated results
                    requests_file = find_latest_requests_file(data_category_path, test_case['Code'])
                    results = get_aggregated_results(requests_file)
                    
                    # Combine test case with results
                    combined_data = test_case.copy()
                    if results:
                        combined_data.update(results)
                    else:
                        print(f"Warning: No results found for {test_case['Code']}")
                        # Add empty result columns
                        result_columns = [
                            'Total_Requests', 'Failure_Count', 'Median_Response_Time', 
                            'Average_Response_Time', 'Min_Response_Time', 'Max_Response_Time',
                            'Average_Content_Size', 'Requests_per_sec', 'Failures_per_sec',
                            'Response_Time_50%', 'Response_Time_66%', 'Response_Time_75%',
                            'Response_Time_80%', 'Response_Time_90%', 'Response_Time_95%',
                            'Response_Time_98%', 'Response_Time_99%', 'Response_Time_99.9%',
                            'Response_Time_99.99%', 'Response_Time_100%'
                        ]
                        for col in result_columns:
                            combined_data[col] = 'N/A'
                    
                    all_test_cases.append(combined_data)
                    
            except Exception as e:
                print(f"Error processing {tests_json_path}: {e}")
        
        # Write CSV file for this category
        if all_test_cases:
            output_file = os.path.join(base_dir, f"scripts/{category_name.lower()}_test_summary.csv")
            
            # Define column order
            columns = [
                'Code', 'Category', 'Users', 'Ramp up / sec', 'QoS', 
                'Message Size (KB)', 'Publish Rate (msg/sec/user)', 'Test Duration (mins)',
                'Total_Requests', 'Failure_Count', 'Median_Response_Time',
                'Average_Response_Time', 'Min_Response_Time', 'Max_Response_Time',
                'Average_Content_Size', 'Requests_per_sec', 'Failures_per_sec',
                'Response_Time_50%', 'Response_Time_66%', 'Response_Time_75%',
                'Response_Time_80%', 'Response_Time_90%', 'Response_Time_95%',
                'Response_Time_98%', 'Response_Time_99%', 'Response_Time_99.9%',
                'Response_Time_99.99%', 'Response_Time_100%'
            ]
            
            try:
                with open(output_file, 'w', newline='', encoding='utf-8') as csvfile:
                    writer = csv.DictWriter(csvfile, fieldnames=columns)
                    writer.writeheader()
                    writer.writerows(all_test_cases)
                
                print(f"Created: {output_file} with {len(all_test_cases)} test cases")
                
            except Exception as e:
                print(f"Error writing {output_file}: {e}")

def main():
    """Main function to execute the script."""
    print("MQTT Test Cases Summary Generator")
    print("=" * 40)
    
    process_test_cases()
    
    print("\nSummary generation completed!")
    print("\nGenerated files:")
    base_dir = "/app"
    for file in glob.glob(os.path.join(base_dir, "*_test_summary.csv")):
        print(f"  - {os.path.basename(file)}")

if __name__ == "__main__":
    main()
