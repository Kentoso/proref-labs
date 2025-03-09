#include <string>
#include <iostream>
#include <fstream>

using namespace std;

enum class DataType {
    INT,
    FLOAT,
    DOUBLE,
    CHAR,
    STRING
};

struct Attribute {
    int id;
    string name;
    DataType type;
    int size;
    Attribute* next;

    bool isPrimary;
    bool isForeignKey;


    static int GetTypeSize(DataType type) {
        switch (type) {
            case DataType::INT: return sizeof(int);
            case DataType::FLOAT: return sizeof(float);
            case DataType::DOUBLE: return sizeof(double);
            case DataType::CHAR: return sizeof(char);
            case DataType::STRING: return 16; // Default size for string (can be modified as needed)
            default: return 0;
        }
    }

    static Attribute* Create(int id, const string& name, DataType type, Attribute* next = nullptr, bool isPrimary = false, bool isForeignKey = false) {
        int size = GetTypeSize(type);
        return new Attribute{id, name, type, size, next, isPrimary, isForeignKey};
    }
};

struct Type {
    string name;
    int rowSize;
    Attribute* attributes;

    static Type* CreateType(string name, Attribute* attributes) {
        Type* type = new Type;
        type->name = name;
        type->attributes = attributes;
        type->rowSize = 0;
        Attribute* current = attributes;
        while (current != nullptr) {
            type->rowSize += current->size;
            current = current->next;
        }
        return type;
    }
};

struct TableAttributeDefinition {
    string nameFromType;
    int idFromType;
    bool isPrimary;
    bool isForeignKey;

    static vector<TableAttributeDefinition> FromType(Type* type) {
        vector<TableAttributeDefinition> attributes;
        Attribute* current = type->attributes;
        while (current != nullptr) {
            attributes.push_back({current->name, current->id, current->isPrimary, current->isForeignKey});
            current = current->next;
        }
        return attributes;
    }
};

struct Table {
    string tableName;
    string fileName;
    Type* type;
    vector<TableAttributeDefinition> attributes;
    Table* next;
};

struct TableListElement {
    Type* type;
    Table* homogeneousTablesHead;
    TableListElement* next;
};

struct TableList {
    TableListElement* head;

    void print() const {
        const TableListElement* element = head;
        int elementIndex = 0;
        while (element) {
            std::cout << "TableListElement " << elementIndex << ":\n";
            if (element->type) {
                std::cout << "  Type:\n";
                std::cout << "    Row Size: " << element->type->rowSize << "\n";
                std::cout << "    Attributes:\n";
                const Attribute* attr = element->type->attributes;
                while (attr) {
                    std::cout << "      Attribute ID: " << attr->id << "\n";
                    std::cout << "      Name: " << attr->name << "\n";
                    std::cout << "      Type: ";
                    switch (attr->type) {
                        case DataType::INT: std::cout << "INT"; break;
                        case DataType::FLOAT: std::cout << "FLOAT"; break;
                        case DataType::DOUBLE: std::cout << "DOUBLE"; break;
                        case DataType::CHAR: std::cout << "CHAR"; break;
                        case DataType::STRING: std::cout << "STRING"; break;
                    }
                    std::cout << "\n";
                    std::cout << "      Size: " << attr->size << "\n";
                    std::cout << "      isPrimary: " << (attr->isPrimary ? "true" : "false") << "\n";
                    std::cout << "      isForeignKey: " << (attr->isForeignKey ? "true" : "false") << "\n";
                    attr = attr->next;
                }
            }
            if (element->homogeneousTablesHead) {
                std::cout << "  Tables:\n";
                const Table* table = element->homogeneousTablesHead;
                int tableIndex = 0;
                while (table) {
                    std::cout << "    Table " << tableIndex << ":\n";
                    std::cout << "      Table Type Size: " << table->type->rowSize << "\n";
                    std::cout << "      Table Name: " << table->tableName << "\n";
                    std::cout << "      File Name: " << table->fileName << "\n";
                    std::cout << "      Table Attributes:\n";
                    for (size_t i = 0; i < table->attributes.size(); i++) {
                        const TableAttributeDefinition& tad = table->attributes[i];
                        std::cout << "        Attribute " << i << ":\n";
                        std::cout << "          Name from Type: " << tad.nameFromType << "\n";
                        std::cout << "          ID from Type: " << tad.idFromType << "\n";
                        std::cout << "          isPrimary: " << (tad.isPrimary ? "true" : "false") << "\n";
                        std::cout << "          isForeignKey: " << (tad.isForeignKey ? "true" : "false") << "\n";
                    }
                    table = table->next;
                    tableIndex++;
                }
            }
            element = element->next;
            elementIndex++;
        }
    }

    void FInsertR(Table* newTable) {
        // check if newTable has type which is already in the list
        TableListElement* element = head;
        while (element) {
            if (element->type == newTable->type) {
                newTable->next = element->homogeneousTablesHead;
                element->homogeneousTablesHead = newTable;
                return;
            }
            element = element->next;
        }

        // if newTable has a new type, create a new TableListElement
        TableListElement* newElement = new TableListElement{newTable->type, newTable, head};
        head = newElement;
    }

    void commit(const string& filename) const {
        // open file in binary write mode
        ofstream ofs(filename, ios::binary);
        if (!ofs) {
            cerr << "Error: cannot open file " << filename << " for writing table list.\n";
            return;
        }

        // write the number of elements in the list
        int numElements = 0;
        const TableListElement* element = head;
        while (element) {
            numElements = numElements + 1;
            element = element->next;
        }

        // write everything as bytes
        ofs.write(reinterpret_cast<const char*>(&numElements), sizeof(numElements));
        element = head;

        // write each element
        int currIndex = 0;
        while (element) {
            // index
            ofs.write(reinterpret_cast<const char*>(&currIndex), sizeof(currIndex));
            // length of the type name
            int typeNameLen = element->type->name.size();
            ofs.write(reinterpret_cast<const char*>(&typeNameLen), sizeof(typeNameLen));
            // type name
            ofs.write(element->type->name.data(), typeNameLen);
            // row size
            ofs.write(reinterpret_cast<const char*>(&element->type->rowSize), sizeof(element->type->rowSize));
            // number of attributes
            int numAttributes = 0;
            Attribute* attr = element->type->attributes;
            while (attr) {
                numAttributes = numAttributes + 1;
                attr = attr->next;
            }
            ofs.write(reinterpret_cast<const char*>(&numAttributes), sizeof(numAttributes));
            // write each attribute
            attr = element->type->attributes;
            while (attr) {
                // attribute id
                ofs.write(reinterpret_cast<const char*>(&attr->id), sizeof(attr->id));
                // length of the attribute name
                int attrNameLen = attr->name.size();
                ofs.write(reinterpret_cast<const char*>(&attrNameLen), sizeof(attrNameLen));
                // attribute name
                ofs.write(attr->name.data(), attrNameLen);
                // attribute type
                int attrType = static_cast<int>(attr->type);
                ofs.write(reinterpret_cast<const char*>(&attrType), sizeof(attrType));
                // attribute size
                ofs.write(reinterpret_cast<const char*>(&attr->size), sizeof(attr->size));
                // isPrimary
                int isPrimary = attr->isPrimary ? 1 : 0;
                ofs.write(reinterpret_cast<const char*>(&isPrimary), sizeof(isPrimary));
                // isForeignKey
                int isForeignKey = attr->isForeignKey ? 1 : 0;
                ofs.write(reinterpret_cast<const char*>(&isForeignKey), sizeof(isForeignKey));
                attr = attr->next;
            }
            // number of tables
            int numTables = 0;
            Table* table = element->homogeneousTablesHead;
            while (table) {
                numTables = numTables + 1;
                table = table->next;
            }
            ofs.write(reinterpret_cast<const char*>(&numTables), sizeof(numTables));
            // write each table
            table = element->homogeneousTablesHead;
            while (table) {
                // table name
                int tableNameLen = table->tableName.size();
                ofs.write(reinterpret_cast<const char*>(&tableNameLen), sizeof(tableNameLen));
                ofs.write(table->tableName.data(), tableNameLen);
                // file name
                int fileNameLen = table->fileName.size();
                ofs.write(reinterpret_cast<const char*>(&fileNameLen), sizeof(fileNameLen));
                ofs.write(table->fileName.data(), fileNameLen);
                // number of attributes
                int numTableAttributes = table->attributes.size();
                ofs.write(reinterpret_cast<const char*>(&numTableAttributes), sizeof(numTableAttributes));
                // write each table attribute
                for (const auto& tad : table->attributes) {
                    // name from type
                    int nameFromTypeLen = tad.nameFromType.size();
                    ofs.write(reinterpret_cast<const char*>(&nameFromTypeLen), sizeof(nameFromTypeLen));
                    ofs.write(tad.nameFromType.data(), nameFromTypeLen);
                    // id from type
                    ofs.write(reinterpret_cast<const char*>(&tad.idFromType), sizeof(tad.idFromType));
                    // isPrimary
                    int isPrimary = tad.isPrimary ? 1 : 0;
                    ofs.write(reinterpret_cast<const char*>(&isPrimary), sizeof(isPrimary));
                    // isForeignKey
                    int isForeignKey = tad.isForeignKey ? 1 : 0;
                    ofs.write(reinterpret_cast<const char*>(&isForeignKey), sizeof(isForeignKey));
                }
                table = table->next;
            }
            
            element = element->next;
            currIndex = currIndex + 1;
        }
    }

    static TableList* deserialize(const string& filename) {
        ifstream ifs(filename, ios::binary);
        if (!ifs) {
            cerr << "Error: cannot open file " << filename << " for reading table list.\n";
            return nullptr;
        }

        // Read number of elements in the list
        int numElements = 0;
        ifs.read(reinterpret_cast<char*>(&numElements), sizeof(numElements));

        vector<TableListElement*> elements;

        for (int i = 0; i < numElements; i++) {
            // (a) Read the stored index (not used in reconstruction)
            int index;
            ifs.read(reinterpret_cast<char*>(&index), sizeof(index));

            // (b) Read type name length and type name
            int typeNameLen = 0;
            ifs.read(reinterpret_cast<char*>(&typeNameLen), sizeof(typeNameLen));
            string typeName;
            typeName.resize(typeNameLen);
            ifs.read(&typeName[0], typeNameLen);

            // (c) Read rowSize for the type
            int rowSize = 0;
            ifs.read(reinterpret_cast<char*>(&rowSize), sizeof(rowSize));

            // (d) Read number of attributes in the type
            int numAttributes = 0;
            ifs.read(reinterpret_cast<char*>(&numAttributes), sizeof(numAttributes));

            // Reconstruct the linked list of Attributes
            Attribute* attrHead = nullptr;
            Attribute* attrTail = nullptr;
            for (int j = 0; j < numAttributes; j++) {
                int attrId;
                ifs.read(reinterpret_cast<char*>(&attrId), sizeof(attrId));

                int attrNameLen = 0;
                ifs.read(reinterpret_cast<char*>(&attrNameLen), sizeof(attrNameLen));
                string attrName;
                attrName.resize(attrNameLen);
                ifs.read(&attrName[0], attrNameLen);

                int attrTypeInt = 0;
                ifs.read(reinterpret_cast<char*>(&attrTypeInt), sizeof(attrTypeInt));
                DataType attrType = static_cast<DataType>(attrTypeInt);

                int attrSize = 0;
                ifs.read(reinterpret_cast<char*>(&attrSize), sizeof(attrSize));

                int isPrimaryInt = 0;
                ifs.read(reinterpret_cast<char*>(&isPrimaryInt), sizeof(isPrimaryInt));
                bool isPrimary = (isPrimaryInt == 1);

                int isForeignKeyInt = 0;
                ifs.read(reinterpret_cast<char*>(&isForeignKeyInt), sizeof(isForeignKeyInt));
                bool isForeignKey = (isForeignKeyInt == 1);

                // Allocate and initialize new Attribute
                Attribute* newAttr = new Attribute;
                newAttr->id = attrId;
                newAttr->name = attrName;
                newAttr->type = attrType;
                newAttr->size = attrSize;
                newAttr->isPrimary = isPrimary;
                newAttr->isForeignKey = isForeignKey;
                newAttr->next = nullptr;

                if (!attrHead) {
                    attrHead = newAttr;
                    attrTail = newAttr;
                } else {
                    attrTail->next = newAttr;
                    attrTail = newAttr;
                }
            }

            // Create the Type instance; assign rowSize from file
            Type* type = new Type;
            type->name = typeName;
            type->attributes = attrHead;
            type->rowSize = rowSize;

            // (e) Read number of tables for this type element
            int numTables = 0;
            ifs.read(reinterpret_cast<char*>(&numTables), sizeof(numTables));

            // Reconstruct the linked list of Tables
            Table* tablesHead = nullptr;
            Table* tablesTail = nullptr;
            for (int k = 0; k < numTables; k++) {
                // Table name
                int tableNameLen = 0;
                ifs.read(reinterpret_cast<char*>(&tableNameLen), sizeof(tableNameLen));
                string tableName;
                tableName.resize(tableNameLen);
                ifs.read(&tableName[0], tableNameLen);

                // File name
                int fileNameLen = 0;
                ifs.read(reinterpret_cast<char*>(&fileNameLen), sizeof(fileNameLen));
                string fileName;
                fileName.resize(fileNameLen);
                ifs.read(&fileName[0], fileNameLen);

                // Number of table attribute definitions
                int numTableAttributes = 0;
                ifs.read(reinterpret_cast<char*>(&numTableAttributes), sizeof(numTableAttributes));
                vector<TableAttributeDefinition> tableAttrDefs;
                for (int l = 0; l < numTableAttributes; l++) {
                    int nameFromTypeLen = 0;
                    ifs.read(reinterpret_cast<char*>(&nameFromTypeLen), sizeof(nameFromTypeLen));
                    string nameFromType;
                    nameFromType.resize(nameFromTypeLen);
                    ifs.read(&nameFromType[0], nameFromTypeLen);

                    int idFromType = 0;
                    ifs.read(reinterpret_cast<char*>(&idFromType), sizeof(idFromType));

                    int isPrimaryInt = 0;
                    ifs.read(reinterpret_cast<char*>(&isPrimaryInt), sizeof(isPrimaryInt));
                    bool tadPrimary = (isPrimaryInt == 1);

                    int isForeignKeyInt = 0;
                    ifs.read(reinterpret_cast<char*>(&isForeignKeyInt), sizeof(isForeignKeyInt));
                    bool tadForeignKey = (isForeignKeyInt == 1);

                    TableAttributeDefinition tad;
                    tad.nameFromType = nameFromType;
                    tad.idFromType = idFromType;
                    tad.isPrimary = tadPrimary;
                    tad.isForeignKey = tadForeignKey;
                    tableAttrDefs.push_back(tad);
                }

                // Create and initialize Table
                Table* newTable = new Table;
                newTable->tableName = tableName;
                newTable->fileName = fileName;
                newTable->type = type;  // Points to the type we just reconstructed
                newTable->attributes = tableAttrDefs;
                newTable->next = nullptr;

                if (!tablesHead) {
                    tablesHead = newTable;
                    tablesTail = newTable;
                } else {
                    tablesTail->next = newTable;
                    tablesTail = newTable;
                }
            }

            // Create the TableListElement for this type and its tables
            TableListElement* newElement = new TableListElement;
            newElement->type = type;
            newElement->homogeneousTablesHead = tablesHead;
            newElement->next = nullptr;
            elements.push_back(newElement);
        }

        // Chain all TableListElement objects in order
        TableListElement* listHead = nullptr;
        TableListElement* listTail = nullptr;
        for (auto elem : elements) {
            if (!listHead) {
                listHead = elem;
                listTail = elem;
            } else {
                listTail->next = elem;
                listTail = elem;
            }
        }

        TableList* tableList = new TableList;
        tableList->head = listHead;
        return tableList;
    }
};

void test1() {
        cout << "----------------------------------------" << endl;
    cout << "Init table list with the Customer table" << endl;
    cout << "----------------------------------------" << endl;

    Attribute* attr1 = Attribute::Create(1, "CustomerID", DataType::INT, nullptr, true);
    Attribute* attr2 = Attribute::Create(2, "CustomerName", DataType::STRING, attr1);
    
    Type* customerType = Type::CreateType("customer", attr2);

    vector<TableAttributeDefinition> customerAttributes1 = TableAttributeDefinition::FromType(customerType);
    vector<TableAttributeDefinition> customerAttributes2 = TableAttributeDefinition::FromType(customerType);

    Table* customerTable1 = new Table{"Customers1", "customers1.txt", customerType, customerAttributes1, nullptr};
    Table* customerTable2 = new Table{"Customers2", "customers2.txt", customerType, customerAttributes2, customerTable1};

    TableListElement* customerElement = new TableListElement{customerType, customerTable2, nullptr};

    TableList* tableList = new TableList{customerElement};

    tableList->print();

    cout << "----------------------------------------" << endl;
    cout << "Adding Hotel table to the list using FInsertR" << endl;
    cout << "----------------------------------------" << endl;

    Attribute *newAttr1 = Attribute::Create(1, "HotelID", DataType::INT, nullptr, true);
    Attribute *newAttr2 = Attribute::Create(2, "HotelName", DataType::STRING, newAttr1);
    Attribute *newAttr3 = Attribute::Create(3, "HotelLocation", DataType::STRING, newAttr2);

    Type *hotelType = Type::CreateType("hotel", newAttr3);

    vector<TableAttributeDefinition> hotelAttributes = TableAttributeDefinition::FromType(hotelType);

    Table *hotelTable = new Table{"Hotels", "hotels.txt", hotelType, hotelAttributes, nullptr};

    tableList->FInsertR(hotelTable);

    tableList->print();
    
    tableList->commit("tableList.bin");
}

void test2() {
    cout << "----------------------------------------" << endl;
    cout << "Deserializing table list from file" << endl;
    cout << "----------------------------------------" << endl;

    TableList* tableList = TableList::deserialize("tableList.bin");
    tableList->print();
}

int main() {
    test1();
    // test2();

    return 0;
}
