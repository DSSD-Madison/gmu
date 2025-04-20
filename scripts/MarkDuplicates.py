import subprocess
from collections import defaultdict


#HOW TO USE THIS FILE
# 1.) first run migrate_db script so you have an up to date copy of prod data on local db
# 2.) look at main method below

def main():
    #this method takes in data from local db and returns a list of uuid's where the coresponding document
    #has a title ( extracted from s3_file link ) that appears in other s_3 file names
    #All but one of the "copies" are a part of this list
    #when choosing a copy to keep, it prioritizes titled documents
    listToMark = getUUIDToMark()
    #this method just writes the uuid's from the list into duplicate_uuids.txt which is used if you want to update prod
    #if theres already a duplicate_uuids.txt file with data, delete it before running this method
    write_uuids_to_file_for_prod_update(listToMark)
    #if you want to update local database duplicates, then use this, second parameter is batch size
    update_local(listToMark,100)

    #NOW you can update production if you choose too. duplicate_uuids.txt should have updated uuid's to mark, so you can
    #run mark_prod_duplciates.sh which uses and updates uuid's from that list as duplicates


#this method returns a list where each value represents a line from the output of selecting id, s3_file, and title from documents
#and is used in getUUIDToMark()
def GetDataFromLocal():
        # Just the SQL query now — no need for \c
    query = "SELECT id, s3_file, title FROM documents;"

    # Connect directly to mydatabase
    command = [
        "docker", "exec", "-i", "mypostgres",
        "psql", "-U", "postgres", "-d", "mydatabase",  # <- specify your target DB here
        "-c", query
    ]
    # Run the command
    result = subprocess.run(command, capture_output=True, text=True)
    # Output
    if result.returncode == 0:
        output = result.stdout
        outputList = output.split('\n')
    return outputList

def getUUIDToMark():
    #get raw data to work with
    outputList = GetDataFromLocal()
    title_uuid_dict = defaultdict(list)
    #trust me with the ranges here, first 2 and last 3 lines contain no relevant info
    for index in range(2,len(outputList)-3):
        #get different data from each line of the output, creates a the super useful dict
        line = outputList[index]
        parts = line.split('|')
        uuid = parts[0].strip()
        title = parts[2].strip()
        s3_file_name = parts[1].strip()
        s_3title = s3_file_name.split('/')[-1]
        #after this dict is populated, it has the key as the real title extracted from s3_file link, and the value
        #is a list where each value is the uuid with three @'s and then the title from the actual title field
        title_uuid_dict[s_3title].append(str(uuid) +"@@@"+str(title))

    list_uuid_to_mark = []
    for key in title_uuid_dict:
        key_list = title_uuid_dict[key]
        #if theres multiple files, look through each and count how many are untitled, if they all are, pick one to not mark as dup
        if len(key_list) > 1:
            #keep track of uuid of items with untitled in them, we
            #would prefer to keep these marked
            untitled_count = sum(1 for item in key_list if "Untitled" in item)
            if untitled_count == len(key_list):
                #if all are untitled, just keep the first, rest store to mark as dup
                for i in range(1,len(key_list)):
                    list_uuid_to_mark.append(key_list[i].split('@@@')[0])

            else:
                #find first index where it's titled, track that, then add uuid from every index besides
                #that one
                first_index_with_title = 0
                for i in range(len(key_list)):
                    if "Untitled" not in key_list[i]:
                        first_index_with_title = i
                for i in range(len(key_list)):
                    if i != first_index_with_title:
                        list_uuid_to_mark.append(title_uuid_dict[key][i].split('@@@')[0])
    #clean
    list_uuid_to_mark = [uuid for uuid in list_uuid_to_mark if uuid.strip()]
    return list_uuid_to_mark



def update_local(list_uuid_to_mark, batch_size):
    def chunks(lst, batch_size):
        for i in range(0, len(lst), batch_size):
            yield lst[i:i + batch_size]

    for batch in chunks(list_uuid_to_mark, batch_size):
        # Escape single quotes and format UUIDs for SQL
        uuid_list = ", ".join(f"'{uuid}'" for uuid in batch)
        query = f"UPDATE documents SET has_duplicate = TRUE WHERE id IN ({uuid_list});"

        command = [
            "docker", "exec", "-i", "mypostgres",
            "psql", "-U", "postgres", "-d", "mydatabase",
            "-c", query
        ]

        result = subprocess.run(command, capture_output=True, text=True)

        if result.returncode != 0:
            print(f"[ERROR] Batch failed:\n{result.stderr.strip()}")
        else:
            print(f"[OK] Batch of {len(batch)} UUIDs updated.")



#method for updating uuid's file with our found uuid's to be used in the update prod script
def write_uuids_to_file_for_prod_update(uuids, filepath="duplicate_uuids.txt"):
    # Remove empty and whitespace-only entries
    cleaned = [uuid.strip() for uuid in uuids if uuid.strip()]
    with open(filepath, "w") as f:
        for uuid in cleaned:
            f.write(uuid + "\n")


#MISC NOTES / EXAMPLE to check duplicates were marked
#shown below is two example titles that appears in multiple s3 buckets
#and is the s3_file pertaining to each uuid
#looking at the s3_file names is confirmation that the code is working as intended
#If before running script, all has_duplicate tags for these documents were false, after
#running script, for each, 1 should be marked false as has_dup, others shoul dbe marked as true
#
#first one - rodlarsen-remarks-bahrainopen.pdf
#53490296-3c1d-4d36-aecd-8407150aef5c -> s3://ipinst-org/English/rodlarsen-remarks-bahrainopen.pdf
#b519e90c-59fc-4a73-b8f7-a1b3b60af4ec -> s3://ipinst-org-issue-briefs/english/rodlarsen-remarks-bahrainopen.pdf
#fdbaa4c2-e7b7-4f28-a7c0-aa7f0de8125e -> s3://ipinst-org-policy-papers/rodlarsen-remarks-bahrainopen.pdf

#second one - LOCALIZING-THE-2030-AGENDA-SYNTHESIS.pdf
#7b7bba3d-fdb7-4658-9beb-03e70a78d99e -> s3://ipinst-org/English/LOCALIZING-THE-2030-AGENDA-SYNTHESIS.pdf
#e68be571-f1fd-4cd0-958d-35dc538fae99 -> s3://ipinst-org-issue-briefs/english/LOCALIZING-THE-2030-AGENDA-SYNTHESIS.pdf
#f2b761f2-a2fb-4d62-90f8-a3f72b86fd60 -> s3://ipinst-org-policy-papers/LOCALIZING-THE-2030-AGENDA-SYNTHESIS.pdf


#CHECK ABOVE EXAMPLES VALUES IN LOCAL DB AFTER UPDATING LOCAL DB

uuids1 = [
    "53490296-3c1d-4d36-aecd-8407150aef5c",
    "b519e90c-59fc-4a73-b8f7-a1b3b60af4ec",
    "fdbaa4c2-e7b7-4f28-a7c0-aa7f0de8125e"]
uuids2 = [
    "7b7bba3d-fdb7-4658-9beb-03e70a78d99e",
    "e68be571-f1fd-4cd0-958d-35dc538fae99",
    "f2b761f2-a2fb-4d62-90f8-a3f72b86fd60"
]

if __name__ == "__main__":
    main()