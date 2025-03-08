#!/usr/bin/env python3

import pandas as pd
import matplotlib.pyplot as plt
import argparse

# Parse arguments
parser = argparse.ArgumentParser(description='Plot size and height vs. generation by type from a CSV file.')
parser.add_argument('input_file', type=str, help='Path to the CSV file')
parser.add_argument('output_file', type=str, help='Path to save the output plot')
args = parser.parse_args()

# Load the data
df = pd.read_csv(args.input_file, names=['generation', 'type', 'size', 'height'])

# Convert size to MB
df['size'] = df['size'] / (1024 * 1024)

# Create figure and axis
fig, ax1 = plt.subplots(figsize=(10, 6))
fig.patch.set_facecolor('white')
ax1.set_facecolor('white')

# Remove top and right borders
for spine in ['top', 'right']:
    ax1.spines[spine].set_visible(False)

# Reduce margins
plt.subplots_adjust(left=0.1, right=0.9, top=0.9, bottom=0.1)

# Plot size on primary y-axis
for t in df['type'].unique():
    subset = df[df['type'] == t]
    ax1.plot(subset['generation'], subset['size'], marker='o', markersize=2, linestyle='-', linewidth=1, label=f"Size - {t}", alpha=0.5)

ax1.set_xlabel('Generation')
ax1.set_ylabel('Size (MB)', color='tab:blue')
ax1.tick_params(axis='y', labelcolor='tab:blue')
ax1.grid(True)

# Create secondary y-axis for height
ax2 = ax1.twinx()
ax2.set_facecolor('white')
for t in df['type'].unique():
    subset = df[df['type'] == t]
    ax2.plot(subset['generation'], subset['height'], marker='', markersize=4, linestyle='--', linewidth=1, color='green', label=f"Height - {t}", alpha=0.5)

ax2.set_ylabel('Height', color='green')
ax2.tick_params(axis='y', labelcolor='green')

# Combine legends
handles1, labels1 = ax1.get_legend_handles_labels()
handles2, labels2 = ax2.get_legend_handles_labels()
ax1.legend(handles1 + handles2, labels1 + labels2, title='Legend', loc='upper left')

# Save plot
plt.title('Size and Height vs. Generation by Type')
plt.savefig(args.output_file)
plt.close()