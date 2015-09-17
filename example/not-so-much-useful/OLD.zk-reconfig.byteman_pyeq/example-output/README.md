# Example Output


## CSV
See csv.csv
You can use gnuplot/LibreOffice/Excel to plot this.


## Execution Pattern

`cat history/$N/json | jq .`

JQ: http://stedolan.github.io/jq/


NOTE: pattern abstraction needs elimination of "uninteresting" arguments(e.g., session ids, timestamps) from FunctionCallEvent.
this elimination can be done in zk.btm by constructing suitable HashMap<String argName, Object argValue>.

