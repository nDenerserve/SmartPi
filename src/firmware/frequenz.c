This file is part of SmartPi.

    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.

    Diese Datei ist Teil von SmartPi.

    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.

    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.

    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
    
    
    

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <linux/i2c-dev.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <time.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <net/if.h>
#include <arpa/inet.h>


char device_address;

#define BUFSIZE 4

char value, phase, phase_write, number;




//int main(void)
int main( int argc, char* argv[] )
{
 char buffer[200];
  char frequenz_1[10];
  char frequenz_2[10];
  char frequenz_3[10]; 

 int v=0, t=0;




FILE *fp;
int temp;

fp = fopen("/run/SmartPi/Smartpi_Value.txt", "r");



if(fp == NULL) 
{
	printf("Datei konnte nicht geoeffnet werden.\n");
}
else 
{
	// komplette Datei zeichenweise ausgeben
	while((temp = fgetc(fp))!=EOF) 
	{
		
		//printf("%c ", temp);
		buffer[v]=temp;
		v++;
		
		
		
		

	}
}

fclose(fp);
v=0;


char delimiter[] = ":";//Trennzeichen (:) zwischen den die Werte stehen, insgesamt 16 Werte!!!
char *ptr;

// initialisieren und ersten Abschnitt erstellen
ptr = strtok(buffer, delimiter);

while(ptr != NULL) 
{
	//printf("Abschnitt gefunden: %s\n", ptr);
	switch(v) 
	{
		case 13: strcpy(frequenz_1, ptr); break;
		case 14: strcpy(frequenz_2, ptr); break;
		case 15: strcpy(frequenz_3, ptr); break;
		default:  break;
	} 
	//strcpy(current_1, ptr);
	v++;
	// naechsten Abschnitt erstellen
 	ptr = strtok(NULL, delimiter);
} 
 

value=atoi(argv[2]);
//printf("\r\nEingabe Value %i\r\n",value);
phase=atoi(argv[3]);
//printf("\r\nEingabe Phase %i\r\n",phase);
		
    	
//-------------------------------------------------------------------------------------

//IP-SCAN
int fd;
 struct ifreq ifr;

 fd = socket(AF_INET, SOCK_DGRAM, 0);

 /* I want to get an IPv4 IP address */
 ifr.ifr_addr.sa_family = AF_INET;

 /* I want IP address attached to "eth0" */
 strncpy(ifr.ifr_name, "eth0", IFNAMSIZ-1);

 ioctl(fd, SIOCGIFADDR, &ifr);

 close(fd);

 /* display result */
 //printf("%s\n", inet_ntoa(((struct sockaddr_in *)&ifr.ifr_addr)->sin_addr));






//TIME
/*time_t rawtime;
  struct tm * timeinfo;

  time ( &rawtime );
  timeinfo = localtime ( &rawtime );
  printf ( "Current local time and date: %s", asctime (timeinfo) );
*/

time_t w = time(NULL);
struct tm tm = *localtime(&w);



//JSON
printf("{");
printf("\"serial\": \"12345678\",");
//printf("\"time\": \"2016-03-11 16:06:00\",");
printf("\"time\": \"%d-%d-%d %d:%d:%d\",", tm.tm_year + 1900, tm.tm_mon + 1, tm.tm_mday, tm.tm_hour, tm.tm_min, tm.tm_sec);
printf("\"softwareversion\": \"1.0.1\",");
printf("\"ipaddress\": \"%s\",",inet_ntoa(((struct sockaddr_in *)&ifr.ifr_addr)->sin_addr));
printf("\"datasets\": [{");
//printf("\"time\": \"2016-03-11 11:20:00\",");
printf("\"time\": \"%d-%d-%d %d:%d:%d\",", tm.tm_year + 1900, tm.tm_mon + 1, tm.tm_mday, tm.tm_hour, tm.tm_min, tm.tm_sec);
printf("\"phases\": [{");




if (phase!=77)number=3;//1 Phase
else number=0;//All
		
for (t=number;t<4;t++)
{
	if(phase==77)
	{
	   	phase_write=t+1;
	}
	else phase_write=phase;

	switch(phase_write)
	{
		
		//Phase 1
		case 1:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Frequenz:
				case 50:
				{
						

					
						printf("\"type\": \"frequenzy\",");
						printf("\"unity\": \"Hz\",");
						printf("\"data\": %s",frequenz_1);
						if(phase==77)printf("}]");
						if(phase==77)printf("}, {");

					

					break;
				

		  	 	}//Ende case 50 frequenzy (phase 1)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 1 (Phase 1)












		//Phase 2
		case 2:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Frequenz:
				case 50:
				{
						

					
						printf("\"type\": \"frequenzy\",");
						printf("\"unity\": \"Hz\",");
						printf("\"data\": %s",frequenz_2);
						if(phase==77)printf("}]");
						if(phase==77)printf("}, {");

					

					break;
				

		  	 	}//Ende case 50 frequenzy (phase 2)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 2 (Phase 2)










		//Phase 3
		case 3:
		{
			
			printf("\"phase\": %i,",phase_write);
			printf("\"name\": \"phase %i\",",phase_write);
			printf("\"values\": [{");
 		
			switch (value)
			{
				//Frequenz:
				case 50:
				{
						
					
						printf("\"type\": \"frequenzy\",");
						printf("\"unity\": \"Hz\",");
						printf("\"data\": %s",frequenz_3);
						//if(phase==77)printf("}]");
						//if(phase==77)printf("}, {");

					

					break;
				

		  	 	}//Ende case 50 frequenzy (phase 3)




	 		}//Ende switch(value)
		
			break;
		}//Ende Case 3 (Phase 3)







		


		


		



		default:
  		{
   	 		//Wenn keine Wahl zutrifft!!!
   	 		break;
  		}



	}//Ende der Switch-Anweisung (PHASE)

}//Ende der for-Schleife
printf("}]");
printf("}]");
printf("}]");
printf("}");

  
  return 0;
}
