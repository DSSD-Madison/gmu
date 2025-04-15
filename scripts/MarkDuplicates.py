import subprocess

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

#now we have each line of the return in form of string in a list
#so we need to find the title from the s3_file name, ( not the one returned ) keep track of uuid's as well

#I think the best way to do this would be creatig a dictionary with the key as the title derived from the s3_file
#and then the vales will be a list of uuid's.

#This way we can figure out how many real duplicates there are ( not just by actual title )
#by keepig count of how many key's have values of length greater than one


#first two lines and last three don't contain any relevant information
#which is why we use the range
from collections import defaultdict

title_uuid_dict = defaultdict(list)
for index in range(2,len(outputList)-3):
    line = outputList[index]
    parts = line.split('|')
    uuid = parts[0].strip()
    title = parts[2].strip()
    s3_file_name = parts[1].strip()
    s_3title = s3_file_name.split('/')[-1]
    title_uuid_dict[s_3title].append(str(uuid) +"@@@"+str(title))
#There are 165 instances where title in Untitled, and theres no other s3_files names with that title in it
#There uuids are stored in untitled_no_dup_uuid, we could go in later and update the title field of these documents
#to reflect the end of the s3_file path ( aka real name )

#now, let's figure out which uuid's to mark

list_uuid_to_mark = []
total_count = 0
case_where_one_is = 0
flag = False
for key in title_uuid_dict:
    key_list = title_uuid_dict[key]
    if not flag:
        print(key)
        print(key_list)
        flag = True
    #if theres multiple files, look through each and count how many are untitled, if they all are, pick one to not mark as dup
    if len(key_list) > 1:
        total_count +=1
        #keep track of uuid of items with untitled in them, we
        #would prefer to keep these marked
        untitled_count = sum(1 for item in key_list if "Untitled" in item)
        if untitled_count == len(key_list):
            #if all are untitled, just keep the first, rest store to mark as dup
            for i in range(1,len(key_list)):
                list_uuid_to_mark.append(key_list[i].split('@@@')[0])

        else:
            case_where_one_is  += 1
            #find first index where it's titled, track that, then add uuid from every index besides
            #that one
            first_index_with_title = 0
            for i in range(len(key_list)):
                if "Untitled" not in key_list[i]:
                    first_index_with_title = i
            for i in range(len(key_list)):
                if i != first_index_with_title:
                    list_uuid_to_mark.append(title_uuid_dict[key][i].split('@@@')[0])

print(str(case_where_one_is))
#we can do one last check on our list of uuid's to mark
#it's length should be equal to the sum of all the length of all the lists - 1 where the length is >1

theoretical_count = 0
for key in title_uuid_dict:
    key_list = title_uuid_dict[key]
    #if theres multiple files, look through each and count how many are untitled, if they all are, pick one randomley to not mark as dup
    if len(key_list) > 1:
        theoretical_count += (len(key_list) - 1)

print(list_uuid_to_mark)

if theoretical_count == len(list_uuid_to_mark):
    print("Yay, counts match, should be good to remove uuids in list")
    print("theoretical = "+ str(theoretical_count) +" and real is "+ str(len(list_uuid_to_mark)))
else:
    print("They Don't Match, theoretical = "+ str(theoretical_count) +" and real is "+ str(len(list_uuid_to_mark)))
#clean
list_uuid_to_mark = [uuid for uuid in list_uuid_to_mark if uuid.strip()]

#things seem to match, so now we have a list of uuid's where we can mark their document's is_dup as true;
# summary of how we got here
# - for each document, get "real title" from s3_file field - the last part of the link
# - store that title as a key, and the value's represent uuid's where we extracted that title
# - if theres multiple documents with the same extracted title, try to find one with the title field as not "untitled" not mark as dup
# - if we can't find one, just choose first one to not mark as dup

#LETS FINALLY UPDATE IN LOCAL DB, or at least make the methods
def chunks(lst, batch_size):
    for i in range(0, len(lst), batch_size):
        yield lst[i:i + batch_size]

def update_local(list_uuid_to_mark, batch_size):
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

#write_uuids_to_file_for_prod_update(list_uuid_to_mark)
#update_local(list_uuid_to_mark,100)





#MISC NOTES / EXAMPLE to check duplicates were marked
#shown below is two example titles that appears in multiple s3 buckets
#and is the s3_file pertaining to each uuid
#looking at the s3_file names is confirmation that the code is working as intended
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

uuids = [
    "53490296-3c1d-4d36-aecd-8407150aef5c",
    "b519e90c-59fc-4a73-b8f7-a1b3b60af4ec",
    "fdbaa4c2-e7b7-4f28-a7c0-aa7f0de8125e",
    "7b7bba3d-fdb7-4658-9beb-03e70a78d99e",
    "e68be571-f1fd-4cd0-958d-35dc538fae99",
    "f2b761f2-a2fb-4d62-90f8-a3f72b86fd60"
]

for uuid in uuids:
    query = f"SELECT has_duplicate FROM documents WHERE id = '{uuid}';"
    command = [
        "docker", "exec", "-i", "mypostgres",
        "psql", "-U", "postgres", "-d", "mydatabase",
        "-t", "-c", query
    ]
    result = subprocess.run(command, capture_output=True, text=True)
    output = result.stdout.strip()
    print(f"{uuid}: {output}")

#LOOKS LIKE IT WORKED ( AT LEAST THE FIRST TIME ) !!
#kinda, I wanted to verify before and after so I migrated prod over to local
#and then all the values in local were true for has_dup, so ....