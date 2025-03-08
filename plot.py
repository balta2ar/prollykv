#!/usr/bin/env python3

import pandas as pd
import matplotlib.pyplot as plt
import argparse

# Parse arguments
parser = argparse.ArgumentParser(description='Plot size vs. generation by type from a CSV file.')
parser.add_argument('input_file', type=str, help='Path to the CSV file')
parser.add_argument('output_file', type=str, help='Path to save the output plot')
args = parser.parse_args()

# Load the data
df = pd.read_csv(args.input_file, names=['generation', 'type', 'size'])

# Convert size to MB
df['size'] = df['size'] / (1024 * 1024)

# Plot the data
plt.figure(figsize=(10, 6))
for t in df['type'].unique():
    subset = df[df['type'] == t]
    plt.plot(subset['generation'], subset['size'], marker='o', linestyle='-', label=t)

plt.xlabel('Generation')
plt.ylabel('Size (MB)')
plt.title('Size vs. Generation by Type')
plt.legend(title='Type')
plt.grid(True)

# Save plot
plt.savefig(args.output_file)
plt.close()
