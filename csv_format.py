import pandas as pd
import os
import numpy as np

input_file = os.path.expandvars(
    "$HOME/Downloads/3000000to3999999_BlockTransaction/3000000to3999999_BlockTransaction.csv"
)
output_file = "output.csv"

usecols = ["from", "to", "toCreate", "fromIsContract", "toIsContract", "value", "gasPrice", "gasUsed"]
df = pd.read_csv(input_file, usecols=usecols, nrows=500000)

df = df.replace("None", np.nan)

mask_toCreate = df["toCreate"].isna()
mask_from = df["fromIsContract"] == 0
mask_to = df["toIsContract"] == 0

mask = mask_toCreate & mask_from & mask_to
df = df[mask]

for i in range(3):
    df.insert(0, f"blank{i}", "")

df.insert(df.columns.get_loc("value")+1, "blank_after_value", "")

for i in range(6):
    df[f"blank_end{i}"] = ""

df.iloc[:, 1] = np.arange(1, len(df) + 1)

df.to_csv(output_file, index=False, header=False)

print(f"Remaining Number of Data: {len(df)}")
