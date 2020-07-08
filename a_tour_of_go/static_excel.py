import os
import sys
import xlrd
import re

total_map = {}

def ReadExcel(filename, sheetnum, begin, end):
    wb = xlrd.open_workbook(file_name)#打开文件

    print (wb)

    sheet = wb.sheet_by_index(sheetnum - 1)

    for i in range(1, sheet.nrows):
        row_list = sheet.row_values(i) # 每一行的数据在row_list 数组里
        for cell_str in row_list:
            cell_str.strip()
            if len(cell_str) == 0:
                continue
            cell_str_list = re.split("、| |\n|\t", cell_str)
            for word in cell_str_list:
                word.strip()
                if len(word) <= 1 or len(word) > 4 :
                    continue
                if word in total_map:
                    total_map[word] = total_map[word] + 1
                else:
                    total_map[word] = 1
    for key,value in total_map.items():
        print ("{0} : {1}".format(key,value))
            


if __name__ == "__main__":
    file_name = "教师值日表3.xlsx"
    sheetnum = 1
    begin = "B6"
    end = "I11"
    ReadExcel(file_name, sheetnum, begin, end)


