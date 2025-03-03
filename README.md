# OpenIntel Parquet Downloader

## 📌 Overview  
This software is designed to **automate the download of all Parquet files** from the following directory:  
🔗 [https://openintel.nl/download/forward-dns/basis%3Dtoplist/](https://openintel.nl/download/forward-dns/basis%3Dtoplist/)  

## 🚀 Features  
- **Batch download**: Fetch all available Parquet files from the directory.
- **Efficient handling**: Optimized for high-performance downloading and storage management.

## 🛠️ Installation  
To use this software, ensure you have the following dependencies installed:  

### **Prerequisites**  
- Golang 1.23.5 

### **Install**  
```sh
go install github.com/gustavorobertux/gopenintel@latest

$ gopeintel -h

Usage of gopenintel:
  -end-year int
    	End year (maximum 2025) (default 2025)
  -help
    	Display help menu
  -proxy string
    	HTTP proxy URL (optional)
  -start-year int
    	Start year (minimum 2016) (default 2016)
```
### Example
```sh
gopenintel -start-year 2024 -end-year 2025
```

### **Suggested Usage**
For optimal use, you should have a Parquet file reader. In my case, I used DuckDB.

See below an example of usage.

```sh
$ snap install duckdb ( Ubuntu )
 
$ duckdb -c "COPY (SELECT DISTINCT query_name FROM parquet_scan('part-00000-053e7dcd-88f7-4938-8911-8a38ae169f71-c000.gz.parquet') WHERE query_name LIKE '%att.com%' LIMIT 1000000) TO 'output.csv' (FORMAT CSV, HEADER FALSE);" && cat output.csv
```
